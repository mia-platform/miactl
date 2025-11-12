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
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCreateVersionListCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := VersionListCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestListProjectVersionsCmd(t *testing.T) {
	testCases := map[string]struct {
		options          listVersionsOptions
		expectError      bool
		expectedErrorMsg string
		testServer       *httptest.Server
		outputTextJSON   string
	}{
		"error missing project id": {
			options:          listVersionsOptions{},
			expectError:      true,
			expectedErrorMsg: "missing project ID, please provide a project ID with the --project-id flag",
			testServer: versionListTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"valid versions list with json output": {
			options: listVersionsOptions{
				ProjectID:    "test-project",
				OutputFormat: "json",
			},
			testServer: versionListTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`[
						{"tagName": "v1.0.0", "ref": "main", "message": "First version", "releaseDescription": "Initial release"},
						{"tagName": "v1.1.0", "ref": "main", "message": "Second version", "releaseDescription": "Feature update"}
					]`))
					return true
				}
				return false
			}),
			outputTextJSON: `[
				{"tagName": "v1.0.0", "ref": "main", "message": "First version", "releaseDescription": "Initial release"},
				{"tagName": "v1.1.0", "ref": "main", "message": "Second version", "releaseDescription": "Feature update"}
			]`,
		},
		"valid versions list with yaml output": {
			options: listVersionsOptions{
				ProjectID:    "test-project",
				OutputFormat: "yaml",
			},
			testServer: versionListTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`[
						{"tagName": "v1.0.0", "ref": "main", "message": "First version", "releaseDescription": "Initial release"},
						{"tagName": "v1.1.0", "ref": "main", "message": "Second version", "releaseDescription": "Feature update"}
					]`))
					return true
				}
				return false
			}),
			outputTextJSON: `[
				{"tagName": "v1.0.0", "ref": "main", "message": "First version", "releaseDescription": "Initial release"},
				{"tagName": "v1.1.0", "ref": "main", "message": "Second version", "releaseDescription": "Feature update"}
			]`,
		},
		"empty versions list": {
			options: listVersionsOptions{
				ProjectID:    "test-project",
				OutputFormat: "json",
			},
			testServer: versionListTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`[]`))
					return true
				}
				return false
			}),
			outputTextJSON: `No versions found for the project`,
		},
		"server error": {
			options: listVersionsOptions{
				ProjectID:    "test-project",
				OutputFormat: "json",
			},
			expectError:      true,
			expectedErrorMsg: "bad request",
			testServer: versionListTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"statusCode": 400, "message": "bad request"}`))
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

			outputBuffer := bytes.NewBuffer([]byte{})

			err = listProjectVersions(ctx, client, testCase.options, outputBuffer)

			if testCase.expectError {
				require.Error(t, err)
				require.EqualError(t, err, testCase.expectedErrorMsg)
			} else {
				require.NoError(t, err)

				if testCase.options.OutputFormat == encoding.JSON && testCase.outputTextJSON != "No versions found for the project" {
					found := outputBuffer.String()
					require.JSONEq(t, testCase.outputTextJSON, found, "output should match expected JSON")
				} else if testCase.options.OutputFormat == encoding.YAML && testCase.outputTextJSON != "No versions found for the project" {
					foundMap := []map[string]interface{}{}
					err := yaml.Unmarshal(outputBuffer.Bytes(), &foundMap)
					require.NoError(t, err)

					expectedMap := []map[string]interface{}{}
					err = json.Unmarshal([]byte(testCase.outputTextJSON), &expectedMap)
					require.NoError(t, err)

					require.Equal(t, expectedMap, foundMap)
				} else if testCase.outputTextJSON == "No versions found for the project" {
					require.Contains(t, outputBuffer.String(), "No versions found for the project")
				}
			}
		})
	}
}

func versionListTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler(w, r) {
			return
		}

		t.Logf("unexpected request: %#v\n%#v", r.URL, r)
		w.WriteHeader(http.StatusNotFound)
		assert.Fail(t, "unexpected request")
	}))
}
