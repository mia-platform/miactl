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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		authCacheProvider client.AuthCacheReadWriter
		expectedToken     string
		testServer        *httptest.Server
	}{
		"valid jwt access token": {
			authCacheProvider: &testAuthCacheProvider{},
			expectedToken:     "foo",
			testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				assert.Fail(t, "we don't expect to call the test server")
			}),
		},
		"expired jwt access": {
			authCacheProvider: &testAuthCacheProvider{expired: true},
			expectedToken:     "new",
			testServer:        testServerForServiceAccount(t),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.testServer.Close()
			restConfig := &client.Config{
				Host:      testCase.testServer.URL,
				Transport: http.DefaultTransport,
			}

			restClient, err := client.APIClientForConfig(restConfig)
			require.NoError(t, err)
			ua := &serviceAccountAuthenticator{
				userAuth:     testCase.authCacheProvider,
				client:       restClient,
				clientID:     "id",
				clientSecret: "secret",
			}

			token, err := ua.AccessToken()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func TestJWTAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		authCacheProvider client.AuthCacheReadWriter
		expectedToken     string
		testServer        *httptest.Server
	}{
		"valid jwt access token": {
			authCacheProvider: &testAuthCacheProvider{},
			expectedToken:     "foo",
			testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				assert.Fail(t, "we don't expect to call the test server")
			}),
		},
		"expired jwt access": {
			authCacheProvider: &testAuthCacheProvider{expired: true},
			expectedToken:     "new",
			testServer:        testServerForServiceAccount(t),
		},
	}

	keyData := testKeyData(t)
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.testServer.Close()
			restConfig := &client.Config{
				Host:      testCase.testServer.URL,
				Transport: http.DefaultTransport,
			}

			restClient, err := client.APIClientForConfig(restConfig)
			require.NoError(t, err)
			ua := &serviceAccountAuthenticator{
				userAuth:       testCase.authCacheProvider,
				client:         restClient,
				clientID:       "id",
				keyID:          "miactl",
				privateKeyData: keyData,
			}

			token, err := ua.AccessToken()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func TestBasicAuthenticatorRoundTrip(t *testing.T) {
	rt := &testRoundTripper{}
	auth := &serviceAccountAuthenticator{
		userAuth: &testAuthCacheProvider{},
		client:   nil,
		next:     rt,
	}

	req := &http.Request{
		Header: make(http.Header),
	}

	auth.RoundTrip(req) //nolint:bodyclose
	rtRequest := rt.Request
	require.NotNil(t, rtRequest)
	assert.NotSame(t, rtRequest, req)
	assert.Equal(t, "Bearer foo", rtRequest.Header.Get("Authorization"))

	req = &http.Request{
		Header: make(http.Header),
	}
	req.Header.Set("Authorization", "Bearer existing")
	req.Body = io.NopCloser(bytes.NewBuffer([]byte("")))

	auth.RoundTrip(req) //nolint:bodyclose
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
