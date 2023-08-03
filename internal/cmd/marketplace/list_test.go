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

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestListMarketplaceItems(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companiesURI string
		err          bool
	}{
		"valid get response": {
			server:       mockServer(t, true),
			companiesURI: listMarketplaceEndpoint,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
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
			err = listMarketplaceItems(client, "")
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockServer(t *testing.T, validResponse bool) *httptest.Server {
	t.Helper()
	validBodyString := `[
	{
		"_id": "63775c07a09ac0996ebfb7ef",
		"name": "ACL Service",
		"description": "This service provides REST API that can be configured to manage request ACL",
		"type": "plugin",
		"imageUrl": "/v2/files/download/5354b282-d5eb-4089-91f1-b5108a1e98f4.png",
		"supportedByImageUrl": "/v2/files/download/83b11dd9-41f6-4920-bb2d-2a809e944851.png",
		"supportedBy": "Mia-Platform",
		"category": {
			"id": "auth",
			"label": "Core Plugins - Authentication & Authorization"
		},
		"repositoryUrl": "https://git.tools.mia-platform.eu/platform/core/acl-service",
		"componentsIds": [],
		"publishOnMiaDocumentation": true,
		"documentation": {
			"type": "externalLink",
			"url": "https://docs.mia-platform.eu/docs/runtime_suite/acl-service/overview"
		}
	}
]`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != listMarketplaceEndpoint && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(http.StatusOK)
		if validResponse {
			w.Write([]byte(validBodyString))
			return
		}
		w.Write([]byte("invalid json"))
	}))
}
