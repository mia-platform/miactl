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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestBuildMarketplaceItemsList(t *testing.T) {
	testCases := map[string]struct {
		options          GetMarketplaceItemsOptions
		server           *httptest.Server
		clientConfig     *client.Config
		err              bool
		expectedContains []string
	}{
		"private company marketplace": {
			options: GetMarketplaceItemsOptions{
				companyID: "my-company",
				public:    false,
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
					t.FailNow()
				case strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
					r.Method == http.MethodGet &&
					r.URL.Query().Get("tenantId") == "my-company":
					_, err := w.Write([]byte(marketplacePrivateCompanyBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE", "COMPANY ID",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin", "my-company",
			},
		},
		"wrong payload": {
			options: GetMarketplaceItemsOptions{
				companyID: "my-company",
				public:    false,
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
					r.Method == http.MethodGet &&
					r.URL.Query().Get("tenantId") == "my-company":
					_, err := w.Write([]byte("{}"))
					require.NoError(t, err)
				}
			})),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err:              true,
			expectedContains: []string{},
		},
		"public marketplace": {
			options: GetMarketplaceItemsOptions{
				public: true,
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
					r.Method == http.MethodGet &&
					!r.URL.Query().Has("tenantId"):
					_, err := w.Write([]byte(marketplaceItemsBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE", "COMPANY ID",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin", "my-company",
				"43774c07d09ac6996ecfb3eg", "a-public-service", "A public service", "plugin", "another-company",
			},
		},
		"should retrieve public marketplace when company and public are being set": {
			options: GetMarketplaceItemsOptions{
				companyID: "my-company",
				public:    true,
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
					r.Method == http.MethodGet &&
					!r.URL.Query().Has("tenantId"):
					_, err := w.Write([]byte(marketplaceItemsBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE", "COMPANY ID",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin", "my-company",
				"43774c07d09ac6996ecfb3eg", "a-public-service", "A public service", "plugin", "another-company",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			found, err := getMarketplaceItemsTable(context.TODO(), client, testCase.options)
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

func marketplacePrivateCompanyBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"_id": "43774c07d09ac6996ecfb3ef",
			"name": "Space Travel Service",
			"itemId": "space-travel-service",
			"tenantId": "my-company",
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
}

func marketplaceItemsBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"_id": "43774c07d09ac6996ecfb3ef",
			"name": "Space Travel Service",
			"itemId": "space-travel-service",
			"tenantId": "my-company",
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
		},
		{
			"_id": "43774c07d09ac6996ecfb3eg",
			"name": "A public service",
			"itemId": "a-public-service",
			"tenantId": "another-company",
			"description": "This service provides a REST API to book your next journey to space!",
			"type": "plugin",
			"imageUrl": "/v2/files/download/space.png",
			"supportedByImageUrl": "/v2/files/download/23b12dd9-43a6-4920-cb2d-2a809d946851.png",
			"category": {
				"id": "auth",
				"label": "Core Plugins - Travel"
			},
			"repositoryUrl": "https://git.com/plugins/core/space-travel-service",
			"documentation": {
				"type": "externalLink",
				"url": "https://docs.my-company.eu/docs/space-travel-service/overview"
			},
			"visibility": {
				"public": true
			}
		}
	]`
}
