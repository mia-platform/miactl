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
			testServer:   testServer(t),
			projectID:    "found",
			resourceType: ServicesResourceType,
		},
		"list deployments with empty response": {
			testServer:   testServer(t),
			projectID:    "empty",
			resourceType: DeploymentResourceType,
		},
		"failed request": {
			testServer:   testServer(t),
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

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listEndpointTemplate, "found", "env-id", ServicesResourceType):
			service := resources.Service{
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
			data, err := resources.EncodeResourceToJSON([]resources.Service{service})
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
