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
	"testing"
	"time"

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

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func testServerForCompleteFlow(t *testing.T) *httptest.Server {
	t.Helper()
	accessToken := "new"
	return testServer(t, func(w http.ResponseWriter, r *http.Request) {
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
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"statusCode\":500,\"message\":\"not implemented\"}"))
			assert.Fail(t, "not implemented")
		case r.Method == http.MethodPost && r.RequestURI == refreshTokenEndpointString:
			accessToken = "expired-refresh"
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"statusCode\":500,\"message\":\"not implemented\"}"))
		case r.Method == http.MethodPost && r.RequestURI == getTokenEndpointString:
			testUserToken := resources.UserToken{
				AccessToken:  accessToken,
				RefreshToken: "refresh",
				ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
			}
			encoder := json.NewEncoder(w)
			err := encoder.Encode(&testUserToken)
			require.NoError(t, err)
		default:
			assert.Failf(t, "unexpected request", "%s request %s", r.Method, r.RequestURI)
		}
	})
}

type testRoundTripper struct {
	Request *http.Request
	Err     error
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Request = req
	return nil, rt.Err
}
