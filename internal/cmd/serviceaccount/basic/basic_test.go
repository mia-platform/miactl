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

package basic

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountCmd(t *testing.T) {
	cmd := ServiceAccountCmd(&clioptions.CLIOptions{})
	assert.NotNil(t, cmd)
}

func TestCreateBasicServiceAccount(t *testing.T) {
	testCases := map[string]struct {
		server              *httptest.Server
		serviceAccountName  string
		companyID           string
		role                resources.ServiceAccountRole
		expectedCredentials []string
		expectErr           bool
	}{
		"create a new service account": {
			server:             testServer(t),
			serviceAccountName: "new-sa",
			companyID:          "company",
			role:               resources.ServiceAccountRoleReporter,
			expectedCredentials: []string{
				"client-id",
				"client-secret",
			},
		},
		"server return error": {
			server:              testServer(t),
			serviceAccountName:  "new-sa",
			companyID:           "error",
			role:                resources.ServiceAccountRoleReporter,
			expectErr:           true,
			expectedCredentials: nil,
		},
		"wrong role": {
			server:    testServer(t),
			role:      resources.ServiceAccountRole("wrong"),
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.server
			defer server.Close()
			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)
			credentials, err := createBasicServiceAccount(
				client,
				testCase.serviceAccountName,
				testCase.companyID,
				testCase.role,
			)

			if testCase.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedCredentials, credentials)
		})
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(companyServiceAccountsEndpointTemplate, "company"):
			body := &basicServiceAccountResponse{
				ClientID:         "client-id",
				ClientSecret:     "client-secret",
				ClientIDIssuedAt: 0,
				Company:          "company",
			}
			data, err := resources.EncodeResourceToJSON(body)
			require.NoError(t, err)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(companyServiceAccountsEndpointTemplate, "error"):
			body := &resources.APIError{
				Message:    "error",
				StatusCode: 400,
			}
			data, err := resources.EncodeResourceToJSON(body)
			require.NoError(t, err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))

	return server
}
