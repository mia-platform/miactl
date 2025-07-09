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

package project

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApplyCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ApplyCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestApplyProjectCmd(t *testing.T) {
	testCases := map[string]struct {
		options          applyProjectOptions
		expectError      bool
		expectedErrorMsg string
		testServer       *httptest.Server
		expectedRequest  string
	}{
		"error missing project id": {
			options:          applyProjectOptions{},
			expectError:      true,
			expectedErrorMsg: "missing project name, please provide a project name as argument",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error missing file path": {
			options: applyProjectOptions{
				ProjectID: "test-project",
			},
			expectError:      true,
			expectedErrorMsg: "missing file path, please provide a file path with the -f flag",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error missing revision/version": {
			options: applyProjectOptions{
				ProjectID: "test-project",
				FilePath:  "testdata/valid-config.json",
			},
			expectError:      true,
			expectedErrorMsg: "missing revision/version name, please provide one as argument",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error both revision/version specified": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				VersionName:  "test-version",
				FilePath:     "testdata/valid-config.json",
			},
			expectError:      true,
			expectedErrorMsg: "both revision and version specified, please provide only one",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error invalid file path": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/nonexistent.json",
			},
			expectError:      true,
			expectedErrorMsg: "failed to read project configuration file",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"valid project apply with revision": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/valid-config.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					// Verify the request body structure
					var requestBody map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					// Check that the request has the expected structure
					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.Contains(t, requestBody, "deletedElements")
					assert.Equal(t, "[CLI] Apply project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return true
				}
				return false
			}),
		},
		"valid project apply with version": {
			options: applyProjectOptions{
				ProjectID:   "test-project",
				VersionName: "test-version",
				FilePath:    "testdata/valid-config.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions/test-version/configuration" && r.Method == http.MethodPost {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return true
				}
				return false
			}),
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

			var writer bytes.Buffer
			err = applyProject(context.Background(), client, testCase.options, &writer)

			if testCase.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, writer.String(), "Project configuration applied successfully")
			}
		})
	}
}

func applyTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !handler(w, r) {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
}
