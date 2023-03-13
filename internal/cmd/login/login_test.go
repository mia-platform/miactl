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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalLoginOIDC(t *testing.T) {
	providerID := "the-provider"
	code := "my-code"
	state := "my-state"
	endpoint := "http://127.0.0.1:53534"
	callbackPath := "/api/oauth/token"

	t.Run("correctly returns token", func(t *testing.T) {
		l, err := net.Listen("tcp", ":53534")
		if err != nil {
			panic(err)
		}

		s := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err != nil {
					fmt.Println(err)
				}
				var data map[string]interface{}
				err = json.NewDecoder(r.Body).Decode(&data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				switch {
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && data["code"] == code && data["state"] == state:
					handleCallbackSuccesfulToken(w, r)
					return
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && (data["code"] != code || data["state"] != state):
					handleCallbackUnsuccesfulToken(w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}),
		}
		defer s.Close()

		go s.Serve(l)
		expectedToken := tokens{
			AccessToken:  "accesstoken",
			RefreshToken: "refreshToken",
			ExpiresAt:    23345,
		}

		browser := fakeBrowser{
			code:        code,
			state:       state,
			callbackUrl: callbackUrl,
		}
		tokens, err := GetTokensWithOIDC(endpoint, providerID, browser)
		if err != nil {
			fmt.Println(err)
		}
		require.Equal(t, *tokens, expectedToken)

	})

	t.Run("return error with incorrect callback", func(t *testing.T) {
		l, err := net.Listen("tcp", ":53534")
		if err != nil {
			panic(err)
		}

		s := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err != nil {
					fmt.Println(err)
				}
				var data map[string]interface{}
				err = json.NewDecoder(r.Body).Decode(&data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				switch {
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && data["code"] == code && data["state"] == state:
					handleCallbackSuccesfulToken(w, r)
					return
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && (data["code"] != code || data["state"] != state):
					handleCallbackUnsuccesfulToken(w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}),
		}
		defer s.Close()

		go s.Serve(l)
		callbackUrl := "http://127.0.0.1:45536"
		browser := fakeBrowser{
			code:        code,
			state:       state,
			callbackUrl: callbackUrl,
		}
		_, err = GetTokensWithOIDC(callbackUrl, providerID, browser)
		if err != nil {
			fmt.Println(err)
		}
		require.Error(t, err)
	})

}
func TestOpenBrowser(t *testing.T) {
	t.Run("return error with incorrect provider url", func(t *testing.T) {
		incorrectUrl := "incorrect"
		browser := browser{}
		err := browser.open(incorrectUrl)
		require.Error(t, err)

	})

}

func handleCallbackSuccesfulToken(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{\"AccessToken\":\"accesstoken\", \"RefreshToken\":\"refreshToken\", \"ExpiresAt\":23345}"))
}

func handleCallbackUnsuccesfulToken(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusForbidden)

}
