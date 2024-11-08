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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mia-platform/miactl/client/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestBasicAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		authToken     *oauth2.Token
		expectedToken string
		testServer    *httptest.Server
	}{
		"valid jwt access token": {
			authToken: &oauth2.Token{
				AccessToken:  "foo",
				RefreshToken: "",
				Expiry:       time.Now().Add(1 * time.Hour),
			},
			expectedToken: "foo",
			testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				assert.Fail(t, "we don't expect to call the test server")
			}),
		},
		"expired jwt access": {
			authToken: &oauth2.Token{
				AccessToken:  "foo",
				RefreshToken: "",
				Expiry:       time.Now(),
			},
			expectedToken: "new",
			testServer:    testServerForServiceAccount(t),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.testServer.Close()

			ua, err := NewServiceAccountAuthenticator(&transport.Config{}, AuthConfig{
				Server:       testCase.testServer.URL,
				ClientID:     "id",
				ClientSecret: "secret",
			})
			ua.userAuth = testCase.authToken
			require.NoError(t, err)

			token, err := ua.accessToken()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func TestJWTAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		expectedToken string
		authToken     *oauth2.Token
		testServer    *httptest.Server
	}{
		"valid jwt access token": {
			authToken: &oauth2.Token{
				AccessToken:  "foo",
				RefreshToken: "",
				Expiry:       time.Now().Add(1 * time.Hour),
			},
			expectedToken: "foo",
			testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				assert.Fail(t, "we don't expect to call the test server")
			}),
		},
		"expired jwt access": {
			authToken: &oauth2.Token{
				AccessToken:  "foo",
				RefreshToken: "",
				Expiry:       time.Now(),
			},
			expectedToken: "new",
			testServer:    testServerForServiceAccount(t),
		},
	}

	keyData := testKeyData(t)
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.testServer.Close()

			ua, err := NewServiceAccountAuthenticator(&transport.Config{}, AuthConfig{
				Server:         testCase.testServer.URL,
				ClientID:       "id",
				KeyID:          "miactl",
				PrivateKeyData: keyData,
			})
			ua.userAuth = testCase.authToken
			require.NoError(t, err)

			token, err := ua.accessToken()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func TestBasicAuthenticatorRoundTrip(t *testing.T) {
	rt := &testRoundTripper{}
	auth := &Authenticator{
		userAuth: &oauth2.Token{
			AccessToken:  "foo",
			RefreshToken: "",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
		client: nil,
	}

	req := &http.Request{
		Header: make(http.Header),
	}

	saa := &serviceAccountAuthenticator{
		authenticator: auth,
		next:          rt,
	}
	saa.RoundTrip(req) //nolint:bodyclose
	rtRequest := rt.Request
	require.NotNil(t, rtRequest)
	assert.NotSame(t, rtRequest, req)
	assert.Equal(t, "Bearer foo", rtRequest.Header.Get("Authorization"))

	req = &http.Request{
		Header: make(http.Header),
	}
	req.Header.Set("Authorization", "Bearer existing")
	req.Body = io.NopCloser(bytes.NewBuffer([]byte("")))

	saa.RoundTrip(req) //nolint:bodyclose
	rtRequest = rt.Request
	require.NotNil(t, rtRequest)
	assert.Same(t, rtRequest, req)
	assert.Equal(t, "Bearer existing", rtRequest.Header.Get("Authorization"))
}

func testServerForServiceAccount(t *testing.T) *httptest.Server {
	t.Helper()
	return testServer(t, func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		switch {
		case r.Method == http.MethodPost && r.RequestURI == serviceAccountAuthEndpoint && contentType == formEncoded:
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("{\"access_token\":\"new\",\"token_type\":\"Bearer\",\"expires_in\":3600}"))
		default:
			assert.Failf(t, "unexpected request", "%s request %s", r.Method, r.RequestURI)
		}
	})
}

func testKeyData(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	der := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8,
	})

	return base64.StdEncoding.EncodeToString(der)
}

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}
