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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type testAuthCacheProvider struct {
	expired      bool
	refreshToken string
}

func (ap *testAuthCacheProvider) ReadJWTToken() *oauth2.Token {
	expiry := time.Now().Add(1 * time.Hour)
	if ap.expired {
		// if the expiry is now, is under the skew delta, and so it count as expired
		expiry = time.Now()
	}
	return &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: ap.refreshToken,
		Expiry:       expiry,
	}
}

func (ap *testAuthCacheProvider) WriteJWTToken(_ *oauth2.Token) {}

func TestUserAuthenticator(t *testing.T) {
	testCases := map[string]struct {
		authCacheProvider client.AuthCacheReadWriter
		expectedToken     string
		testServer        *httptest.Server
		testServerHandler http.HandlerFunc
	}{
		// "valid jwt access token": {
		// 	authCacheProvider: &testAuthCacheProvider{},
		// 	expectedToken:     "foo",
		// 	testServer: testServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		// 		assert.Fail(t, "we don't expect to call the test server")
		// 	}),
		// },
		// "expired jwt access, with refresh token": {
		// 	authCacheProvider: &testAuthCacheProvider{expired: true, refreshToken: "foo"},
		// 	expectedToken:     "foo",
		// 	testServer:        testServer(t),
		// },
		// "expired jwt access, with expired refresh token": {
		// 	authCacheProvider: &testAuthCacheProvider{expired: true, refreshToken: "expired"},
		// 	expectedToken:     "foo",
		// 	testServer:        testServer(t),
		// },
		"expired jwt, without refresh token": {
			authCacheProvider: &testAuthCacheProvider{expired: true},
			expectedToken:     "foo",
			testServer: testServer(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.RequestURI == fmt.Sprintf(providerEndpointStringTemplate, appID):
					testProvider := resources.AuthProvider{
						ID:    "foo",
						Label: "Foo",
						Type:  "foo-type",
					}
					encoder := json.NewEncoder(w)
					err := encoder.Encode([]*resources.AuthProvider{&testProvider})
					require.NoError(t, err)
				case r.Method == http.MethodGet && r.RequestURI == authorizeEndpointString:
					assert.Fail(t, "not implemented")
				default:
					assert.Failf(t, "unexpected request", "%s request %s", r.Method, r.RequestURI)
				}
			}),
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
					req.RawQuery = query.Encode()
					resp, err := http.Get(req.String())
					require.NoError(t, err)
					assert.Equal(t, resp.StatusCode, http.StatusOK)
					return nil
				},
			}

			token, err := ua.AccessToken()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedToken, token.AccessToken)
		})
	}
}

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}
