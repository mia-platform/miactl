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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/stretchr/testify/require"
)

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewListCompaniesCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestListCompanies(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		companiesURI string
		expectedErr  string
	}{
		"valid get response": {
			server:       mockServer(t, true),
			companiesURI: companiesURI,
		},
		"invalid body response": {
			server:       mockServer(t, false),
			expectedErr:  "invalid character",
			companiesURI: companiesURI,
		},
		"missing API": {
			server:       mockServer(t, true),
			expectedErr:  "404 Not Found",
			companiesURI: "/not-found-uri",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			miaClient := httphandler.FakeMiaClient(fmt.Sprintf("%s%s", testCase.server.URL, testCase.companiesURI))
			err := listCompanies(miaClient)
			if testCase.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, testCase.expectedErr)
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
		if r.RequestURI != companiesURI || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
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
