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
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintServicesList(t *testing.T) {
	testCases := map[string]struct {
		testServer   *httptest.Server
		resourceType string
		projectID    string
		err          bool
	}{
		"list services with success": {
			testServer:   listResourceTestServer(t),
			projectID:    "found",
			resourceType: ServicesResourceType,
		},
		"list deployments with success": {
			testServer:   listResourceTestServer(t),
			projectID:    "found",
			resourceType: DeploymentsResourceType,
		},
		"list pods with success": {
			testServer:   listResourceTestServer(t),
			projectID:    "found",
			resourceType: PodsResourceType,
		},
		"list cronjobs with success": {
			testServer:   listResourceTestServer(t),
			projectID:    "found",
			resourceType: CronJobsResourceType,
		},
		"list jobs with success": {
			testServer:   listResourceTestServer(t),
			projectID:    "found",
			resourceType: JobsResourceType,
		},
		"list deployments with empty response": {
			testServer:   listResourceTestServer(t),
			projectID:    "empty",
			resourceType: DeploymentResourceType,
		},
		"failed request": {
			testServer:   listResourceTestServer(t),
			projectID:    "fail",
			err:          true,
			resourceType: PodsResourceType,
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

			err = printList(client, testCase.projectID, testCase.resourceType, "env-id")
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func listResourceTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", ServicesResourceType):
			resource := resources.Service{
				Name:      "service-name",
				Type:      "ClusterIP",
				ClusterIP: "127.0.0.1",
				Ports: []resources.Port{
					{
						Name:       "port-name",
						Port:       8000,
						Protocol:   "TCP",
						TargetPort: "8000",
					},
				},
				Age: time.Now().Add(-time.Hour * 24),
			}
			data, err := resources.EncodeResourceToJSON([]resources.Service{resource})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", DeploymentsResourceType):
			resource := resources.Deployment{
				Name:      "deployment-name",
				Ready:     1,
				Replicas:  1,
				Available: 1,
				Age:       time.Now().Add(-time.Hour * 24),
			}
			data, err := resources.EncodeResourceToJSON([]resources.Deployment{resource})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", PodsResourceType):
			resource := resources.Pod{
				Name:   "pod-name",
				Phase:  "running",
				Status: "ok",
				Age:    time.Now(),
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
			data, err := resources.EncodeResourceToJSON([]resources.Pod{resource})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", CronJobsResourceType):
			resource := resources.CronJob{
				Name:         "cronjob-name",
				Suspend:      true,
				Active:       0,
				Schedule:     "* * * * *",
				Age:          time.Now().Add(-time.Hour * 24),
				LastSchedule: time.Now(),
			}
			data, err := resources.EncodeResourceToJSON([]resources.CronJob{resource})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", JobsResourceType):
			resource := resources.Job{
				Name:           "job-name",
				Active:         0,
				Succeeded:      1,
				Failed:         0,
				Age:            time.Now().Add(-time.Hour * 24),
				StartTime:      time.Now().Add(-time.Second * 60),
				CompletionTime: time.Now(),
			}
			data, err := resources.EncodeResourceToJSON([]resources.Job{resource})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "empty", "env-id", DeploymentsResourceType):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "fail", "env-id", PodsResourceType):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}
