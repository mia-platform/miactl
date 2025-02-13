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

package resources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJob(t *testing.T) {
	testCases := map[string]struct {
		testServer  *httptest.Server
		projectID   string
		environment string
		err         bool
	}{
		"create job end with success": {
			testServer:  createJobTestServer(t),
			projectID:   "success",
			environment: "env-id",
		},
		"create job end with error": {
			testServer:  createJobTestServer(t),
			projectID:   "fail",
			environment: "env-id",
			err:         true,
		},
		"fail if no project id": {
			testServer:  createJobTestServer(t),
			projectID:   "",
			environment: "env-id",
			err:         true,
		},
		"fail if environment": {
			testServer:  createJobTestServer(t),
			projectID:   "success",
			environment: "",
			err:         true,
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

			err = createJob(t.Context(), client, testCase.projectID, testCase.environment, "cronjob-name")
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createJobTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "success", "env-id"):
			response := resources.CreateJob{
				JobName: "new-job-name",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "fail", "env-id"):
			response := resources.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Error creating job",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}
