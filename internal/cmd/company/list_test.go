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

package company

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
)

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestListCompanies(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companiesURI string
		err          bool
	}{
		"valid get response": {
			server: mockServer(t, true),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
		},
		"invalid body response": {
			server: mockServer(t, false),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			err = listCompanies(t.Context(), client, &printer.NopPrinter{})
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
		"tenantId": "test-company-id",
		"name": "Test Company ID",
		"environments": [
			{
				"label": "Development",
				"envId": "development",
				"isProduction": true,
				"cluster": {},
				"deploy": {
					"type": "gitlab-ci",
					"providerId": "gitlab"
				}
			}
		],
		"availableNamespaces": [],
		"environmentsVariables": {
			"type": "gitlab",
			"providerId": "gitlab"
		},
		"pipelines": {
			"type": "gitlab-ci"
		},
		"defaultTemplateId": "project-template-id",
		"repository": {
			"providerId": "git-provider-id",
			"basePath": "git/provider/company/path",
			"visibility": "internal"
		}
	}
]`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != listCompaniesEndpoint && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unsupported call")
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
