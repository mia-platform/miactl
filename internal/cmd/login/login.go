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
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/davidebianchi/go-jsonclient"
)

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

const (
	appID       = "miactl"
	callbackUrl = "127.0.0.1:53535"
)

var (
	state string
	code  string
)

func GetTokensWithOIDC(endpoint string, providerID string, b BrowserI) (*Tokens, error) {
	jsonClient, err := jsonclient.New(jsonclient.Options{BaseURL: fmt.Sprintf("%s/api/", endpoint)})
	if err != nil {
		fmt.Printf("%v", "error generating JsonClient")
	}
	callbackPath := "/oauth/callback"
	l, err := net.Listen("tcp", ":53535")
	if err != nil {
		panic(err)
	}

	// Server HTTP request
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == callbackPath && r.Method == http.MethodGet:
				handleCallback(w, r)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	}
	defer s.Close()

	go s.Serve(l) // actually start the server, it manges all buffered requests

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", endpoint, q.Encode())

	//open the browser to contact the provider
	b.open(startURL)

	req, err := jsonClient.NewRequest(http.MethodPost, "oauth/token", map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return &Tokens{}, err
	}

	fmt.Println(jsonClient)

	token := &Tokens{}
	_, err = jsonClient.Do(req, token)
	if err != nil {
		return &Tokens{}, err
	}

	fmt.Println("token", token)
	fmt.Println("error", err)

	fmt.Println(err)
	return token, nil
}

func handleCallback(w http.ResponseWriter, req *http.Request) {
	qs := req.URL.Query()
	code = qs.Get("code")
	state = qs.Get("state")

	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
}
