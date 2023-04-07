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
	"net"
	"net/http"
	"testing"

	"github.com/mia-platform/miactl/internal/browser"
	"github.com/stretchr/testify/require"
)

func TestLocalLoginOIDC(t *testing.T) {
	providerID := "the-provider"
	code := "my-code"
	state := "my-state"
	endpoint := "http://127.0.0.1:53534"
	callbackPath := "/api/oauth/token"
	callbackURL := "localhost:53535"

	l, err := net.Listen("tcp", ":53534")
	if err != nil {
		panic(err)
	}

	browser := browser.NewFakeURLOpener(t, code, state, callbackURL)
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != callbackPath || r.Method != http.MethodPost {
				http.Error(w, "wrong callback or method", http.StatusBadRequest)
			}

			var data map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			switch {
			case data["code"] == code && data["state"] == state:
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{\"AccessToken\":\"accesstoken\", \"RefreshToken\":\"refreshToken\", \"ExpiresAt\":23345}"))
				return
			case data["code"] != code || data["state"] != state:
				w.WriteHeader(http.StatusForbidden)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	}

	go s.Serve(l)
	defer s.Close()

	t.Run("correctly returns token", func(t *testing.T) {
		expectedToken := Tokens{
			AccessToken:  "accesstoken",
			RefreshToken: "refreshToken",
			ExpiresAt:    23345,
		}

		tokens, err := GetTokensWithOIDC(endpoint, providerID, browser)
		require.NoError(t, err)
		require.Equal(t, *tokens, expectedToken)
	})

	t.Run("return error with incorrect callback", func(t *testing.T) {
		tokens, err := GetTokensWithOIDC(callbackURL, providerID, browser)
		require.Error(t, err)
		require.Nil(t, tokens)
	})
}

func TestOpenBrowser(t *testing.T) {
	t.Run("return error with incorrect provider url", func(t *testing.T) {
		incorrectURL := "incorrect"
		browser := browser.NewURLOpener()
		err := browser.Open(incorrectURL)
		require.Error(t, err)
	})
}
