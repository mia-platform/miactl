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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
)

func TestDiscoverOAuthConfig(t *testing.T) {
	t.Run("resource metadata endpoint not found", func(t *testing.T) {
		server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
		defer server.Close()

		apiClient := apiClientForServer(t, server)
		cfg, err := discoverOAuthConfig(t.Context(), apiClient)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("resource metadata returns invalid JSON", func(t *testing.T) {
		server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not-json"))
		})
		defer server.Close()

		apiClient := apiClientForServer(t, server)
		cfg, err := discoverOAuthConfig(t.Context(), apiClient)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("resource metadata has no authorization_servers", func(t *testing.T) {
		server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(protectedResourceMetadata{})
		})
		defer server.Close()

		apiClient := apiClientForServer(t, server)
		cfg, err := discoverOAuthConfig(t.Context(), apiClient)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("OIDC discovery succeeds", func(t *testing.T) {
		// The server acts as both the protected resource and the authorization
		// server. Its own URL is used as the issuer so the OIDC discovery
		// document can reference it self-consistently.
		var serverURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case protectedResourceMetadataPath:
				json.NewEncoder(w).Encode(protectedResourceMetadata{
					AuthorizationServers: []string{serverURL},
				})
			case "/.well-known/openid-configuration":
				// Minimal OIDC discovery document; issuer must match exactly.
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"issuer":                 serverURL,
					"authorization_endpoint": serverURL + "/authorize",
					"token_endpoint":         serverURL + "/token",
				})
			default:
				http.NotFound(w, r)
			}
		}))
		serverURL = server.URL
		defer server.Close()

		apiClient := apiClientForServer(t, server)
		cfg, err := discoverOAuthConfig(t.Context(), apiClient)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, appID, cfg.ClientID)
		assert.Equal(t, serverURL+"/authorize", cfg.Endpoint.AuthURL)
		assert.Equal(t, serverURL+"/token", cfg.Endpoint.TokenURL)
	})

	t.Run("OIDC discovery fails for returned auth server", func(t *testing.T) {
		server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Return a non-reachable authorization server URL.
			json.NewEncoder(w).Encode(protectedResourceMetadata{ //nolint:errcheck
				AuthorizationServers: []string{"http://127.0.0.1:0"},
			})
		})
		defer server.Close()

		apiClient := apiClientForServer(t, server)
		cfg, err := discoverOAuthConfig(t.Context(), apiClient)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

// apiClientForServer returns an API client configured to talk to server.
func apiClientForServer(t *testing.T, server *httptest.Server) client.Interface {
	t.Helper()
	restConfig := &client.Config{
		Host:      server.URL,
		Transport: http.DefaultTransport,
	}
	apiClient, err := client.APIClientForConfig(restConfig)
	require.NoError(t, err)
	return apiClient
}
