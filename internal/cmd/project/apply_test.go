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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
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
		"error missing revision": {
			options: applyProjectOptions{
				ProjectID: "test-project",
				FilePath:  "testdata/valid-config.json",
			},
			expectError:      true,
			expectedErrorMsg: "missing revision name, please provide a revision name",
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
		"error invalid configuration": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/not-valid-configuration.json",
			},
			expectError:      true,
			expectedErrorMsg: "provided configuration is not valid",
			testServer: applyTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"apply base project (JSON)": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/base-config.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.NotContains(t, requestBody, "fastDataConfig")

					assert.Equal(t, "[miactl] Applied project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))

					return true
				}

				return false
			}),
		},
		"apply base project (YAML)": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/base-config.yaml",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.NotContains(t, requestBody, "fastDataConfig")

					assert.Equal(t, "[miactl] Applied project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))

					return true
				}

				return false
			}),
		},
		"apply config with fastdata": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/config-with-fastdata.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.Contains(t, requestBody, "fastDataConfig")

					assert.Equal(t, "[miactl] Applied project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return true
				}
				return false
			}),
		},
		"apply config with extensionsConfig": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/config-with-extensions.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.Contains(t, requestBody, "extensionsConfig")

					assert.Equal(t, "[miactl] Applied project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return true
				}
				return false
			}),
		},
		"apply config with rbacManagerConfig": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/config-with-microfrontendPluginsConfig.json",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "config")
					assert.Contains(t, requestBody, "previousSave")
					assert.Contains(t, requestBody, "title")
					assert.Contains(t, requestBody, "microfrontendPluginsConfig")
					assert.Contains(t, requestBody["microfrontendPluginsConfig"], "rbacManagerConfig")

					assert.Equal(t, "[miactl] Applied project configuration", requestBody["title"])

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return true
				}
				return false
			}),
		},
		"with custom title": {
			options: applyProjectOptions{
				ProjectID:    "test-project",
				RevisionName: "test-revision",
				FilePath:     "testdata/base-config.json",
				Title:        "apply config custom title",
			},
			testServer: applyTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodPost {
					var requestBody map[string]any
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Contains(t, requestBody, "title")
					assert.Equal(t, "apply config custom title", requestBody["title"])

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

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)

			err = handleApplyProjectConfigurationCmd(ctx, client, testCase.options)

			if testCase.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				// assert.Contains(t, writer.String(), "Project configuration applied successfully")
			}
		})
	}
}

func applyTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !handler(w, r) {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request", "%s: %s", r.Method, r.URL.Path)
		}
	}))
}
