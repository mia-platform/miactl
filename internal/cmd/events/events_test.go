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

package events

import (
	"context"
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

func TestPrintEventsList(t *testing.T) {
	testCases := map[string]struct {
		testServer  *httptest.Server
		projectID   string
		environment string
		err         bool
	}{
		"list event with success": {
			testServer:  testServer(t),
			projectID:   "found",
			environment: "env-id",
		},
		"list event with empty response": {
			testServer:  testServer(t),
			projectID:   "empty",
			environment: "env-id",
		},
		"failed request": {
			testServer:  testServer(t),
			projectID:   "fail",
			environment: "env-id",
			err:         true,
		},
		"fail if no project id": {
			testServer:  testServer(t),
			projectID:   "",
			environment: "env-id",
			err:         true,
		},
		"fail if no environment id": {
			testServer:  testServer(t),
			projectID:   "found",
			environment: "",
			err:         true,
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

			err = printEventsList(context.TODO(), client, testCase.projectID, testCase.environment, "resource")
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRowForEvent(t *testing.T) {
	testCases := map[string]struct {
		event       resources.RuntimeEvent
		expectedRow []string
	}{
		"basic event": {
			event: resources.RuntimeEvent{
				Type:      "Warning",
				FirstSeen: time.Now().Add(-time.Hour * 24),
				LastSeen:  time.Now().Add(-time.Hour * 12),
				Message:   "Message test",
				Reason:    "Reason",
				Object:    "Object",
			},
			expectedRow: []string{"12h", "Warning", "Reason", "Object", "Message test"},
		},
		"missing last seend": {
			event: resources.RuntimeEvent{
				Type:      "Warning",
				FirstSeen: time.Now().Add(-time.Hour * 24),
				Message:   "Message test",
				Reason:    "Reason",
				Object:    "Object",
			},
			expectedRow: []string{"24h", "Warning", "Reason", "Object", "Message test"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForEvent(testCase.event))
		})
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(eventsEndpointTemplate, "found", "env-id", "resource"):
			event := resources.RuntimeEvent{
				Type:      "Warning",
				FirstSeen: time.Now().Add(-time.Hour * 24),
				LastSeen:  time.Now().Add(-time.Hour * 12),
				Message:   "Message test",
				Reason:    "Reason",
				Object:    "Resource object definition",
			}
			data, err := resources.EncodeResourceToJSON([]resources.RuntimeEvent{event})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(eventsEndpointTemplate, "fail", "env-id", "resource"):
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(eventsEndpointTemplate, "empty", "env-id", "resource"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}
