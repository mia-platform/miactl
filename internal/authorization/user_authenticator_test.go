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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
)

func TestUserAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		authCacheProvider client.AuthCacheReadWriter
		expectedToken     string
		testServer        *httptest.Server
		testServerHandler http.HandlerFunc
	}{
		"valid jwt access token": {
			authCacheProvider: &testAuthCacheProvider{},
			expectedToken:     "foo",
			testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				assert.Fail(t, "we don't expect to call the test server")
			}),
		},
		"expired jwt access, with refresh token": {
			authCacheProvider: &testAuthCacheProvider{expired: true, refreshToken: "foo"},
			expectedToken:     "refresh",
			testServer: testServer(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && r.RequestURI == refreshTokenEndpointString:
					testUserToken := resources.UserToken{
						AccessToken:  "refresh",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
					}
					encoder := json.NewEncoder(w)
					err := encoder.Encode(&testUserToken)
					require.NoError(t, err)
				default:
					assert.Failf(t, "unexpected request", "%s request %s", r.Method, r.RequestURI)
				}
			}),
		},
		"expired jwt access, with expired refresh token": {
			authCacheProvider: &testAuthCacheProvider{expired: true, refreshToken: "expired"},
			expectedToken:     "expired-refresh",
			testServer:        testServerForCompleteFlow(t),
		},
		"expired jwt, without refresh token": {
			authCacheProvider: &testAuthCacheProvider{expired: true},
			expectedToken:     "new",
			testServer:        testServerForCompleteFlow(t),
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
			ua := &userAuthenticator{
				userAuth: testCase.authCacheProvider,
				client:   restClient,
				serverReadyHandler: func(s string) error {
					query := make(url.Values)
					query.Set("code", "foo")
					query.Set("state", "bar")
					req, err := url.Parse(s)
					require.NoError(t, err)
					req.Path = callbackEndpointString
					req.RawQuery = query.Encode()
					resp, err := http.Get(req.String())
					if resp.Body != nil {
						resp.Body.Close()
					}
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, resp.StatusCode)
					return err
				},
			}

			token, err := ua.AccessToken()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func TestUserAuthenticatorRoundTrip(t *testing.T) {
	rt := &testRoundTripper{}
	auth := &userAuthenticator{
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
