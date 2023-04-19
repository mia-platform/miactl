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

package login

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/internal/browser"
)

type BasicAuthCredentials struct {
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

type JWTCredentials struct {
	Key  string `yaml:"key"`
	Algo string `yaml:"algo"`
}

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// nolint tagliatelle
type M2MAuthInfo struct {
	AuthType  string               `yaml:"type"`
	BasicAuth BasicAuthCredentials `yaml:"basic,omitempty"`
	JWTAuth   JWTCredentials       `yaml:"jwt,omitempty"`
}

const (
	appID         = "miactl"
	callbackURL   = "127.0.0.1:53535"
	m2mOauthRoute = "/api/m2m/oauth/token"
)

var (
	state string
	code  string
)

func GetTokensWithM2MLogin(endpoint string, authInfo M2MAuthInfo) (*Tokens, error) {
	client := &http.Client{}
	var httpReq *http.Request

	loginEndpoint, err := url.JoinPath(endpoint, m2mOauthRoute)
	if err != nil {
		return nil, fmt.Errorf("error building login endpoint: %w", err)
	}

	if authInfo.AuthType == "basic" {
		data := url.Values{}
		data.Set("grant_type", "client_credentials")
		data.Set("audience", "aud1")

		httpReq, err = http.NewRequest("POST", loginEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("error creating login request: %w", err)
		}
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		encodedClientID := base64.StdEncoding.EncodeToString([]byte(authInfo.BasicAuth.ClientID))
		encodedClientSecret := base64.StdEncoding.EncodeToString([]byte(authInfo.BasicAuth.ClientSecret))
		authHeaderValue := fmt.Sprintf("Basic %s:%s", encodedClientID, encodedClientSecret)
		httpReq.Header.Add("Authorization", authHeaderValue)
	} else {
		return nil, fmt.Errorf("JWT authentication for M2M login is still work in progress")
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending login request: %w", err)
	}
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("M2M login failed with status code: %s", resp.Status)
	}

	rawTokens, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading login response body: %w", err)
	}

	resp.Body.Close()

	return convertTokens(rawTokens)
}

func GetTokensWithOIDC(endpoint string, providerID string, b browser.URLOpener) (*Tokens, error) {
	jsonClient, err := jsonclient.New(jsonclient.Options{BaseURL: fmt.Sprintf("%s/api/", endpoint)})
	if err != nil {
		fmt.Printf("%v", "error generating JsonClient")
		return nil, err
	}
	callbackPath := "/oauth/callback"
	l, err := net.Listen("tcp", "127.0.0.1:53535")
	if err != nil {
		panic(err)
	}

	// Start the HTTP server in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == callbackPath && r.Method == http.MethodGet:
				handleCallback(w, r)
				cancel() // Stop the server after the callback is handled
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	}
	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%v", "error starting server")
			cancel()
		}
	}()

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", endpoint, q.Encode())

	err = b.Open(startURL)
	if err != nil {
		return nil, err
	}

	<-ctx.Done()

	err = server.Shutdown(context.Background())
	if err != nil {
		return nil, err
	}

	req, err := jsonClient.NewRequest(http.MethodPost, "oauth/token", map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return &Tokens{}, err
	}

	token := &Tokens{}

	resp, err := jsonClient.Do(req, token)

	if err != nil {
		return &Tokens{}, err
	}

	defer resp.Body.Close()

	return token, nil
}

func handleCallback(w http.ResponseWriter, req *http.Request) {
	response := `<!DOCTYPE html>
<html>
  <body>
    <script>setTimeout(function() { window.close(); }, 1000);</script>
    <center><h1>Login succeeded!</h1></center>
  </body>
</html>
`

	qs := req.URL.Query()
	code = qs.Get("code")
	state = qs.Get("state")

	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
	if err != nil {
		panic(err)
	}
}

func convertTokens(rawM2MTokens []byte) (*Tokens, error) {
	var rawM2MTokensJSON struct {
		AccessToken string `json:"access_token"` // nolint tagliatelle
		TokenType   string `json:"token_type"`   // nolint tagliatelle
		ExpiresIn   int    `json:"expires_in"`   // nolint tagliatelle
	}
	err := json.Unmarshal(rawM2MTokens, &rawM2MTokensJSON)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling m2m tokens: %w", err)
	}

	var tokens Tokens
	tokens.AccessToken = rawM2MTokensJSON.AccessToken
	tokens.ExpiresAt = time.Now().Add(time.Second * time.Duration(rawM2MTokensJSON.ExpiresIn)).Unix()

	return &tokens, nil
}
