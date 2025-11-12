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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVersionCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := VersionCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestVersionProjectCmd(t *testing.T) {
	testCases := map[string]struct {
		options          versionProjectOptions
		expectError      bool
		expectedErrorMsg string
		testServer       *httptest.Server
	}{
		"error missing project id": {
			options:          versionProjectOptions{},
			expectError:      true,
			expectedErrorMsg: "missing project ID, please provide a project ID with the --project-id flag",
			testServer: versionTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error missing tag name": {
			options: versionProjectOptions{
				ProjectID: "test-project",
			},
			expectError:      true,
			expectedErrorMsg: "missing tag name, please provide a tag name with the --tag flag",
			testServer: versionTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error missing revision": {
			options: versionProjectOptions{
				ProjectID: "test-project",
				TagName:   "test-tag",
			},
			expectError:      true,
			expectedErrorMsg: "missing revision, please provide a revision with the --revision flag",
			testServer: versionTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"error missing message": {
			options: versionProjectOptions{
				ProjectID: "test-project",
				TagName:   "test-tag",
				Ref:       "main",
			},
			expectError:      true,
			expectedErrorMsg: "missing message, please provide a message with the --message flag",
			testServer: versionTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
		},
		"valid version creation": {
			options: versionProjectOptions{
				ProjectID:          "test-project",
				TagName:            "test-tag",
				Ref:                "main",
				Message:            "Test message",
				ReleaseDescription: "Test release description",
			},
			testServer: versionTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodPost {
					// Verify request body
					var requestBody map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&requestBody)
					require.NoError(t, err)

					assert.Equal(t, "test-tag", requestBody["tagName"])
					assert.Equal(t, "main", requestBody["ref"])
					assert.Equal(t, "Test message", requestBody["message"])
					assert.Equal(t, "Test release description", requestBody["releaseDescription"])

					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"id": "123", "tagName": "test-tag"}`))
					return true
				}
				return false
			}),
		},
		"server error": {
			options: versionProjectOptions{
				ProjectID:          "test-project",
				TagName:            "test-tag",
				Ref:                "main",
				Message:            "Test message",
				ReleaseDescription: "Test release description",
			},
			expectError:      true,
			expectedErrorMsg: "failed to create project version: bad request",
			testServer: versionTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions" && r.Method == http.MethodPost {
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

			err = handleVersionProjectCmd(ctx, client, testCase.options)

			if testCase.expectError {
				require.Error(t, err)
				require.EqualError(t, err, testCase.expectedErrorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func versionTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
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
