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

package pods

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintPodsList(t *testing.T) {
	testCases := map[string]struct {
		testServer *httptest.Server
		projectID  string
		err        bool
	}{
		"list pod with success": {
			testServer: testServer(t),
			projectID:  "found",
		},
		"list pod with empty response": {
			testServer: testServer(t),
			projectID:  "empty",
		},
		"failed request": {
			testServer: testServer(t),
			projectID:  "fail",
			err:        true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			server := testCase.testServer
			defer server.Close()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)

			err = printPodsList(client, testCase.projectID, "env-id")
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRowForPod(t *testing.T) {
	testCases := map[string]struct {
		pod         resources.Pod
		expectedRow []string
	}{
		"basic pod": {
			pod: resources.Pod{
				Name:      "pod-name",
				Phase:     "running",
				Status:    "ok",
				StartTime: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component", Version: "version"},
				},
				Containers: []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					Status       string `json:"status"`
				}{
					{
						Name:         "container-name",
						Ready:        true,
						RestartCount: 0,
						Status:       "running",
					},
				},
			},
			expectedRow: []string{"Ok", "pod-name", "component:version", "1/1", "Running", "0", "0s"},
		},
		"pod without component": {
			pod: resources.Pod{
				StartTime: time.Now(),
			},
			expectedRow: []string{"", "", "-", "0/0", "", "0", "0s"},
		},
		"pod without component version": {
			pod: resources.Pod{
				StartTime: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component"},
				},
			},
			expectedRow: []string{"", "", "component", "0/0", "", "0", "0s"},
		},
		"pod without component name": {
			pod: resources.Pod{
				StartTime: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Version: "version"},
				},
			},
			expectedRow: []string{"", "", "-", "0/0", "", "0", "0s"},
		},
		"pod without multiple components": {
			pod: resources.Pod{
				StartTime: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component", Version: "version"},
					{Name: "component"},
				},
			},
			expectedRow: []string{"", "", "component:version, component", "0/0", "", "0", "0s"},
		},
		"pod with multiple containers": {
			pod: resources.Pod{
				StartTime: time.Now(),
				Containers: []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					Status       string `json:"status"`
				}{
					{
						Name:         "container-name",
						Ready:        true,
						RestartCount: 3,
						Status:       "running",
					},
					{
						Name:         "container-name2",
						Ready:        false,
						RestartCount: 1,
						Status:       "running",
					},
				},
			},
			expectedRow: []string{"", "", "-", "1/2", "", "4", "0s"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForPod(testCase.pod))
		})
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id"):
			pod := resources.Pod{
				Name:      "pod-name",
				Phase:     "running",
				Status:    "ok",
				StartTime: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component", Version: "version"},
				},
				Containers: []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					Status       string `json:"status"`
				}{
					{
						Name:         "container-name",
						Ready:        true,
						RestartCount: 0,
						Status:       "running",
					},
				},
			}
			data, err := resources.EncodeResourceToJSON([]resources.Pod{pod})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "fail", "env-id"):
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "empty", "env-id"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}
