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

package deploy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddStatus(t *testing.T) {
	sleepDuration = 0

	testCases := map[string]struct {
		server             *httptest.Server
		status             string
		triggerID          string
		projectID          string
		expectErr          bool
		expectedErrMessage string
	}{
		"add status succeed": {
			server:    testAddStatusServer(t),
			triggerID: "trigger-id",
			projectID: "project-id",
		},
		"add status fails": {
			server:             testAddStatusServer(t),
			triggerID:          "fails-bad-request",
			projectID:          "project-id",
			expectErr:          true,
			expectedErrMessage: "some bad request",
		},
		"without project id": {
			server:             testAddStatusServer(t),
			projectID:          "",
			triggerID:          "trigger-id",
			expectErr:          true,
			expectedErrMessage: fmt.Sprintf(deployStatusErrorRequiredTemplate, "projectId"),
		},
		"without trigger id": {
			server:             testAddStatusServer(t),
			projectID:          "project-id",
			triggerID:          "",
			expectErr:          true,
			expectedErrMessage: fmt.Sprintf(deployStatusErrorRequiredTemplate, "triggerId"),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.server
			defer server.Close()
			options := &clioptions.CLIOptions{
				Endpoint:  server.URL,
				TriggerID: testCase.triggerID,
				ProjectID: testCase.projectID,
			}
			err := runAddDeployStatus(context.TODO(), options, testCase.status)
			if testCase.expectErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func testAddStatusServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("request URL: %s\n", r.URL)
		t.Helper()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(deployStatusTriggerEndpointTemplate, "project-id", "trigger-id"):
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(deployStatusTriggerEndpointTemplate, "project-id", "fails-bad-request"):
			w.WriteHeader(http.StatusBadRequest)

			respBody := `{"error": "Bad Request","message": "some bad request"}`
			w.Write([]byte(respBody))
		default:
			w.WriteHeader(http.StatusNotFound)
			require.FailNowf(t, "unknown http request", "request method: %s request URL: %s", r.Method, r.URL)
		}
	}))

	return server
}
