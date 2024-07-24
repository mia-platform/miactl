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
	"path/filepath"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	sleepDuration = 0

	testCases := map[string]struct {
		server    *httptest.Server
		projectID string
		expectErr bool
	}{
		"pipeline succeed": {
			server:    testServer(t),
			projectID: "correct",
		},
		"missing project ID": {
			server:    testServer(t),
			projectID: "",
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.server
			defer server.Close()
			options := &clioptions.CLIOptions{
				Endpoint:     server.URL,
				ProjectID:    testCase.projectID,
				MiactlConfig: filepath.Join(t.TempDir(), "nofile"),
			}
			err := run(context.TODO(), "environmentName", options)
			if testCase.expectErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(deployProjectEndpointTemplate, "correct"):
			data, err := resources.EncodeResourceToJSON(&resources.DeployProject{
				ID:  1,
				URL: "http://example.com",
			})

			require.NoError(t, err)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(pipelineStatusEndpointTemplate, "correct", 1) && r.URL.Query().Get("environment") == "environmentName":
			data, err := resources.EncodeResourceToJSON(&resources.PipelineStatus{
				ID:     1,
				Status: "succeeded",
			})
			require.NoError(t, err)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			require.FailNowf(t, "unknown http request", "request method: %s request URL: %s", r.Method, r.URL)
		}
	}))

	return server
}
