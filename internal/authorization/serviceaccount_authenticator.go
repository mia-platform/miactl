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

package authorization

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/jws"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	serviceAccountAuthEndpoint = "/api/m2m/oauth/token"
	formEncoded                = "application/x-www-form-urlencoded"
)

type serviceAccountAuthenticator struct {
	mutex          sync.Mutex
	userAuth       client.AuthCacheReadWriter
	client         client.Interface
	next           http.RoundTripper
	clientID       string
	clientSecret   string
	keyID          string
	privateKeyData string
}

func (saa *serviceAccountAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTrip(req, saa.next, saa.AccessToken)
}

func (saa *serviceAccountAuthenticator) AccessToken() (*oauth2.Token, error) {
	saa.mutex.Lock()
	defer saa.mutex.Unlock()

	jwt := saa.userAuth.ReadJWTToken()

	if jwt.Valid() {
		return jwt, nil
	}

	return saa.basicAuth()
}

func (saa *serviceAccountAuthenticator) basicAuth() (*oauth2.Token, error) {
	var jwt *oauth2.Token
	var err error
	switch {
	case len(saa.clientID) > 0 && len(saa.clientSecret) > 0:
		jwt, err = getClientCredentialsToken(context.Background(), saa.client, saa.clientID, saa.clientSecret)
	case len(saa.clientID) > 0 && len(saa.keyID) > 0 && len(saa.privateKeyData) > 0:
		var key *rsa.PrivateKey
		key, err = rsaKeyFromBase64(saa.privateKeyData)
		if err != nil {
			break
		}
		jwt, err = getJWTToken(context.Background(), saa.client, saa.keyID, saa.clientID, key)
	default:
		err = fmt.Errorf("inconsistent auth configuration")
	}

	if jwt != nil {
		saa.userAuth.WriteJWTToken(jwt)
	}

	return jwt, err
}

func getClientCredentialsToken(ctx context.Context, apiClient client.Interface, clientID, clientSecret string) (*oauth2.Token, error) {
	req := apiClient.Post().APIPath(serviceAccountAuthEndpoint)
	endpoint := req.URL()

	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     endpoint.String(),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	tokenContext := context.WithValue(ctx, oauth2.HTTPClient, apiClient.HTTPClient())
	return config.Token(tokenContext)
}

func getJWTToken(ctx context.Context, apiClient client.Interface, keyID, clientID string, key *rsa.PrivateKey) (*oauth2.Token, error) {
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

	values := make(url.Values)
	values.Set("grant_type", "client_credentials")
	values.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	values.Set("client_assertion", signedJWS)
	values.Set("client_id", clientID)
	values.Set("token_endpoint_auth_method", "private_key_jwt")
	response, err := apiClient.
		Post().
		APIPath(serviceAccountAuthEndpoint).
		SetHeader("Content-Type", formEncoded).
		Body([]byte(values.Encode())).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	if response.Error() != nil {
		return nil, response.Error()
	}

	type tempToken struct {
		AccessToken string `json:"access_token"` //nolint: tagliatelle
		TokenTipe   string `json:"token_type"`   //nolint: tagliatelle
		ExpiresIn   int    `json:"expires_in"`   //nolint: tagliatelle
	}

	jwtToken := new(tempToken)

	err = response.ParseResponse(jwtToken)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: jwtToken.AccessToken,
		TokenType:   jwtToken.TokenTipe,
		Expiry:      time.Now().Add(time.Duration(jwtToken.ExpiresIn) * time.Second),
	}, nil
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
