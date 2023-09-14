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

package environments

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreconditions(t *testing.T) {
	err := printEnvironments(nil, "company-id", "")
	assert.Error(t, err)

	err = printEnvironments(nil, "", "project-id")
	assert.Error(t, err)

	err = printEnvironments(nil, "", "")
	assert.Error(t, err)
}

func TestListEnvironments(t *testing.T) {
	projectID := "123456abcdef"
	companyID := "company-id"

	testCases := map[string]struct {
		testServer *httptest.Server
		companyID  string
		projectID  string
		err        bool
	}{
		"list environments": {
			testServer: testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.HasPrefix(r.URL.Path, "/api/backend/projects/") && r.Method == http.MethodGet:
					_, err := w.Write([]byte(projectBodyContent(t)))
					require.NoError(t, err)
				case strings.HasPrefix(r.URL.Path, "/api/tenants") && r.Method == http.MethodGet:
					_, err := w.Write([]byte(clusterBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			companyID: companyID,
			projectID: projectID,
		},
		"list empty environments": {
			testServer: testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.HasPrefix(r.URL.Path, "/api/backend/projects/") && r.Method == http.MethodGet:
					_, err := w.Write([]byte(emptyEnvironmentsBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			companyID: companyID,
			projectID: projectID,
		},
		"error in list project call": {
			testServer: testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})),
			companyID: companyID,
			projectID: projectID,
			err:       true,
		},
		"project in different company": {
			testServer: testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.HasPrefix(r.URL.Path, "/api/backend/projects/") && r.Method == http.MethodGet:
					_, err := w.Write([]byte(emptyEnvironmentsBodyContent(t)))
					require.NoError(t, err)
				}
			})),
			companyID: "wrong",
			projectID: projectID,
			err:       true,
		},
		"error in cluster info call": {
			testServer: testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				default:
					w.WriteHeader(http.StatusNotFound)
					assert.Fail(t, fmt.Sprintf("request not expexted %s", r.URL.Path))
				case strings.HasPrefix(r.URL.Path, "/api/backend/projects/") && r.Method == http.MethodGet:
					_, err := w.Write([]byte(projectBodyContent(t)))
					require.NoError(t, err)
				case strings.HasPrefix(r.URL.Path, "/api/tenants") && r.Method == http.MethodGet:
					w.WriteHeader(http.StatusBadRequest)
				}
			})),
			companyID: companyID,
			projectID: projectID,
			err:       true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.testServer
			defer server.Close()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)

			err = printEnvironments(client, testCase.companyID, testCase.projectID)
			switch testCase.err {
			case false:
				require.NoError(t, err)
			default:
				require.Error(t, err)
			}
		})
	}
}

func testServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	return server
}

func projectBodyContent(t *testing.T) string {
	t.Helper()

	return `{
	"_id": "123456abcdef",
	"environments": [
		{
			"label": "Environment Name",
			"cluster": {
				"clusterId": "123456abcdef",
				"namespace": "namespace-name"
			},
			"description": "Project Description",
			"isProduction": true
		},
		{
			"label": "Environment Name 2",
			"cluster": {
				"clusterId": "123456abcdef",
				"namespace": "namespace-name"
			},
			"description": "Project Description 2",
			"isProduction": false
		}
	],
	"name": "Project Name",
	"projectId": "project-id",
	"tenantId": "company-id"
}`
}

func emptyEnvironmentsBodyContent(t *testing.T) string {
	t.Helper()
	return `{
		"_id": "123456abcdef",
		"environments": [],
		"name": "Project Name",
		"projectId": "project-id",
		"tenantId": "company-id"
	}`
}

func clusterBodyContent(t *testing.T) string {
	t.Helper()
	return `{
	"_id": "123456abcdef",
	"clusterId": "cluster-id",
	"distribution": "test",
	"tenantId": "company-id",
	"vendor": "Golang"
}`
}
