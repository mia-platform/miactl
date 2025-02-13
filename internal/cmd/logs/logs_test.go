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

package logs

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogs(t *testing.T) {
	testCases := map[string]struct {
		testServer  *httptest.Server
		projectID   string
		environment string
		podRegex    string
		err         bool
	}{
		"success": {
			testServer:  testServer(t),
			projectID:   "found",
			environment: "env-id",
			podRegex:    "pod",
		},
		"fail": {
			testServer:  testServer(t),
			projectID:   "fail",
			environment: "env-id",
			podRegex:    "pod",
			err:         true,
		},
		"fail parse regex": {
			testServer:  testServer(t),
			projectID:   "found",
			environment: "env-id",
			podRegex:    `^\/(?!\/)(.*?)`,
			err:         true,
		},
		"fail if no project id": {
			testServer:  testServer(t),
			projectID:   "",
			environment: "env-id",
			err:         true,
		},
		"fail if environment": {
			testServer:  testServer(t),
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
			stream, err := getLogs(t.Context(), client, testCase.projectID, testCase.environment, testCase.podRegex, false)
			if testCase.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			defer stream.Close()
			buf := new(bytes.Buffer)
			buf.ReadFrom(stream)
			resultBody := buf.String()
			assert.Equal(t, "stream returned!", resultBody)
		})
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id"):
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			response := []resources.Pod{
				{
					Name: "pod1",
					Containers: []struct {
						Name         string `json:"name"`
						Ready        bool   `json:"ready"`
						RestartCount int    `json:"restartCount"`
						Status       string `json:"status"`
					}{
						{Name: "container1"},
						{Name: "container2"},
					},
				},
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(logsEndpointTemplate, "found", "env-id"):
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "text/html", r.Header.Get("Accept"))
			w.Write([]byte("stream returned!"))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "fail", "env-id"):
			response := resources.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "not found",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusNotFound)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.URL.Path)
		}
	}))
}
