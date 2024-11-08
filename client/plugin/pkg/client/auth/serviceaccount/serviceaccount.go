// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serviceaccount

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	netutil "github.com/mia-platform/miactl/client/net/util"
	"github.com/mia-platform/miactl/client/plugin/pkg/client/auth/serviceaccount/jws"
	"github.com/mia-platform/miactl/client/transport"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	serviceAccountAuthEndpoint = "/api/m2m/oauth/token"
	formEncoded                = "application/x-www-form-urlencoded"
)

type AuthConfig struct {
	Server         string
	ClientID       string
	ClientSecret   string
	KeyID          string
	PrivateKeyData string
}

type Authenticator struct {
	mutex          sync.Mutex
	userAuth       *oauth2.Token
	client         *http.Client
	server         string
	clientID       string
	clientSecret   string
	keyID          string
	privateKeyData string
}

func NewServiceAccountAuthenticator(config *transport.Config, auth AuthConfig) (*Authenticator, error) {
	t, err := transport.New(config)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: t,
	}
	return &Authenticator{
		client:         client,
		server:         auth.Server,
		clientID:       auth.ClientID,
		clientSecret:   auth.ClientSecret,
		keyID:          auth.KeyID,
		privateKeyData: auth.PrivateKeyData,
	}, nil
}

func (a *Authenticator) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &serviceAccountAuthenticator{
		authenticator: a,
		next:          rt,
	}
}

func (a *Authenticator) basicAuth() (*oauth2.Token, error) {
	var jwt *oauth2.Token
	var err error
	switch {
	case len(a.clientID) > 0 && len(a.clientSecret) > 0:
		jwt, err = getClientCredentialsToken(context.Background(), a.server, a.client, a.clientID, a.clientSecret)
	case len(a.clientID) > 0 && len(a.keyID) > 0 && len(a.privateKeyData) > 0:
		var key *rsa.PrivateKey
		key, err = rsaKeyFromBase64(a.privateKeyData)
		if err != nil {
			break
		}
		jwt, err = getJWTToken(context.Background(), a.client, a.server, a.keyID, a.clientID, key)
	default:
		err = fmt.Errorf("inconsistent auth configuration")
	}

	if jwt != nil {
		a.userAuth = jwt
	}

	return jwt, err
}

func (a *Authenticator) accessToken() (*oauth2.Token, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	jwt := a.userAuth
	if jwt != nil && jwt.Valid() {
		return jwt, nil
	}

	return a.basicAuth()
}

type serviceAccountAuthenticator struct {
	authenticator *Authenticator
	next          http.RoundTripper
}

func (saa *serviceAccountAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				req.Body.Close()
			}
		}()
	}

	if len(req.Header.Get("Authorization")) != 0 {
		reqBodyClosed = true
		return saa.next.RoundTrip(req)
	}

	accessToken, err := saa.authenticator.accessToken()
	if err != nil {
		return nil, err
	}

	clonedReq := netutil.CloneRequest(req)
	accessToken.SetAuthHeader(clonedReq)
	reqBodyClosed = true
	return saa.next.RoundTrip(clonedReq)
}

func getClientCredentialsToken(ctx context.Context, server string, client *http.Client, clientID, clientSecret string) (*oauth2.Token, error) {
	endpoint, err := url.JoinPath(server, serviceAccountAuthEndpoint)
	if err != nil {
		return nil, err
	}

	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     endpoint,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	tokenContext := context.WithValue(ctx, oauth2.HTTPClient, client)
	return config.Token(tokenContext)
}

func getJWTToken(ctx context.Context, client *http.Client, server, keyID, clientID string, key *rsa.PrivateKey) (*oauth2.Token, error) {
	jwsHeader := &jws.Header{
		Typ:       "JWT",
		Algorithm: "RS256",
		KeyID:     keyID,
	}

	jwsClaim := &jws.ClaimSet{
		Iss: clientID,
		Sub: clientID,
		Aud: "console-client-credentials",
		PrivateClaims: map[string]interface{}{
			"jti": uuid.New(),
		},
	}

	signedJWS, err := jws.Encode(jwsHeader, jwsClaim, key)
	if err != nil {
		return nil, err
	}

	data := make(url.Values)
	data.Set("grant_type", "client_credentials")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", signedJWS)
	data.Set("client_id", clientID)
	data.Set("token_endpoint_auth_method", "private_key_jwt")

	endpoint, err := url.JoinPath(server, serviceAccountAuthEndpoint)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", formEncoded)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(bodyData)
	}

	return parseToken(bodyData)
}

func rsaKeyFromBase64(base64Data string) (*rsa.PrivateKey, error) {
	keyData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}

	pemData, _ := pem.Decode(keyData)
	key, err := x509.ParsePKCS8PrivateKey(pemData.Bytes)
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("only rsa key are supported")
	}

	return rsaKey, err
}

func parseBody(body []byte, obj interface{}) error {
	err := json.Unmarshal(body, obj)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error during response parsing: %w", err)
	}

	return nil
}

func parseErrorResponse(bodyData []byte) error {
	responseData := make(map[string]any)
	if err := parseBody(bodyData, &responseData); err != nil {
		return err
	}
	return errors.New(responseData["message"].(string))
}

func parseToken(bodyData []byte) (*oauth2.Token, error) {
	type tempToken struct {
		AccessToken string `json:"access_token"` //nolint: tagliatelle
		TokenType   string `json:"token_type"`   //nolint: tagliatelle
		ExpiresIn   int    `json:"expires_in"`   //nolint: tagliatelle
	}

	jwtToken := new(tempToken)
	if err := parseBody(bodyData, &jwtToken); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: jwtToken.AccessToken,
		TokenType:   jwtToken.TokenType,
		Expiry:      time.Now().Add(time.Duration(jwtToken.ExpiresIn) * time.Second),
	}, nil
}
