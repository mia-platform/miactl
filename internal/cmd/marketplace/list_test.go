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

package marketplace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockResponseBody = `[
	{
		"_id": "43774c07d09ac6996ecfb3ef",
		"name": "Space Travel Service",
		"itemId": "space-travel-service",
		"description": "This service provides a REST API to book your next journey to space!",
		"type": "plugin",
		"imageUrl": "/v2/files/download/space.png",
		"supportedByImageUrl": "/v2/files/download/23b12dd9-43a6-4920-cb2d-2a809d946851.png",
		"supportedBy": "My-Company",
		"category": {
			"id": "auth",
			"label": "Core Plugins - Travel"
		},
		"repositoryUrl": "https://git.com/plugins/core/space-travel-service",
		"documentation": {
			"type": "externalLink",
			"url": "https://docs.my-company.eu/docs/space-travel-service/overview"
		}
	}
]`

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestBuildMarketplaceItemsList(t *testing.T) {
	testCases := map[string]struct {
		server           *httptest.Server
		clientConfig     *client.Config
		companiesURI     string
		err              bool
		expectedContains []string
	}{
		"valid get response": {
			server:       mockServer(t, true),
			companiesURI: listMarketplaceEndpoint,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin",
			},
		},
		"invalid body response": {
			server: mockServer(t, false),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err:          true,
			companiesURI: listMarketplaceEndpoint,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			found, err := buildMarketplaceItemsList(client, "my-company")
			if testCase.err {
				assert.Error(t, err)
				assert.Zero(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, found)
				for _, expected := range testCase.expectedContains {
					assert.Contains(t, found, expected)
				}
			}
		})
	}
}

func mockServer(t *testing.T, validResponse bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != listMarketplaceEndpoint && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(http.StatusOK)
		if validResponse {
			w.Write([]byte(mockResponseBody))
			return
		}
		w.Write([]byte(`{"message": "invalid json"}`))
	}))
}
