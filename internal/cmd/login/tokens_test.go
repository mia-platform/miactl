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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/browser"
	"github.com/stretchr/testify/require"
)

const (
	testBaseURL       = "test.url"
	testURLSha        = "b64868a6476817bde1123f534334c2ce78891fcad65c06667acbfdb9007b5dff"
	testTokens        = `{"accessToken":"test_token","refreshToken":"","expiresAt":9999999999}`
	testExpiredTokens = `{"accessToken":"test_token","refreshToken":"","expiresAt":0}`
	invalidJSON       = `invalid_json`
	validCredentials  = `context1:
  type: basic
  basic:
    clientId: id
    clientSecret: secret
context2:
  type: jwt
  jwt:
    key: 123abc
    algo: 123
default:
  type: basic
  basic:
    clientId: default
    clientSecret: default`
)

var (
	testDirPath string
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

func TestGetTokensWithM2MLogin(t *testing.T) {
	server := mockServer(t)
	defer server.Close()

	authInfo := M2MAuthInfo{
		AuthType: "basic",
		BasicAuth: BasicAuthCredentials{
			ClientID:     "id",
			ClientSecret: "secret",
		},
	}

	tokens, err := GetTokensWithM2MLogin(server.URL, authInfo)
	require.NoError(t, err)
	require.Equal(t, "token", tokens.AccessToken)

	authInfo = M2MAuthInfo{
		AuthType: "basic",
		BasicAuth: BasicAuthCredentials{
			ClientID:     "wrong",
			ClientSecret: "wrong",
		},
	}

	tokens, err = GetTokensWithM2MLogin(server.URL, authInfo)
	require.Nil(t, tokens)
	require.ErrorContains(t, err, "401 Unauthorized")
}

func TestOpenBrowser(t *testing.T) {
	t.Run("return error with incorrect provider url", func(t *testing.T) {
		incorrectURL := "incorrect"
		browser := browser.NewURLOpener()
		err := browser.Open(incorrectURL)
		require.Error(t, err)
	})
}

func TestGetURLSha(t *testing.T) {
	sha := getURLSha(testBaseURL)
	require.Equal(t, testURLSha, sha)
}

func TestGetTokensFromFile(t *testing.T) {
	testDirPath = t.TempDir()
	testFilePath := path.Join(testDirPath, testURLSha)

	// valid JSON
	err := os.WriteFile(testFilePath, []byte(testTokens), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := GetTokensFromFile(testBaseURL, testDirPath)
	require.NoError(t, err)
	var expectedTokens Tokens
	err = json.Unmarshal([]byte(testTokens), &expectedTokens)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expectedTokens, *tokens)

	// invalid JSON
	err = os.WriteFile(testFilePath, []byte(invalidJSON), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GetTokensFromFile(testBaseURL, testDirPath)
	require.ErrorContains(t, err, "invalid character")

	// expired token
	err = os.WriteFile(testFilePath, []byte(testExpiredTokens), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GetTokensFromFile(testBaseURL, testDirPath)
	require.ErrorIs(t, err, ErrExpiredToken)
}

func TestWriteTokensToFile(t *testing.T) {
	testDirPath = t.TempDir()
	testFilePath := path.Join(testDirPath, testURLSha)

	var tokens = &Tokens{
		AccessToken:  "test_token",
		RefreshToken: "",
		ExpiresAt:    9999999999,
	}

	err := WriteTokensToFile(testBaseURL, testDirPath, tokens)
	require.NoError(t, err)
	require.FileExists(t, testFilePath)
	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, testTokens, string(fileContent))
}

func mockServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/api/m2m/oauth/token" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data, err := url.ParseQuery(buf.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if data.Get("grant_type") == "client_credentials" {
			if r.Header.Get("Authorization") != "" {
				mockBasicAuth(t, w, r)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}
	}))
}

func mockBasicAuth(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	encodedAuthString := r.Header.Get("Authorization")
	plainAuthString, err := base64.StdEncoding.DecodeString(encodedAuthString[6:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	credentials := strings.Split(string(plainAuthString), ":")
	if credentials[0] == "id" && credentials[1] == "secret" {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"access_token":"token", "token_type":"Bearer", "expires_in":3600}`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
}
