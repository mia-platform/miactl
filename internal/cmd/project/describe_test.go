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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDescribeCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DescribeCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestDescribeProjectArguments(t *testing.T) {
	t.Run("error missing project", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DescribeCmd(opts)

		cmd.SetArgs([]string{"describe"})
		err := cmd.Execute()

		require.EqualError(t, err, "missing project name, please provide a project name as argument")
	})

	t.Run("error missing revision/version", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DescribeCmd(opts)

		cmd.SetArgs([]string{
			"describe",
			"--project-id", "test-project",
		})
		err := cmd.Execute()

		require.EqualError(t, err, "missing revision/version name, please provide one as argument")
	})

	t.Run("error both revision/version specified", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DescribeCmd(opts)

		cmd.SetArgs([]string{
			"describe",
			"--project-id", "test-project",
			"--revision", "test-revision",
			"--version", "test-version",
		})
		err := cmd.Execute()

		require.EqualError(t, err, "both revision and version specified, please provide only one")
	})
}

func TestDescribeProjectCmd(t *testing.T) {
	testCases := map[string]struct {
		options          describeProjectOptions
		revisionName     string
		versionName      string
		expectError      bool
		expectedErrorMsg string
		testServer       *httptest.Server
	}{
		"valid project with revision": {
			options: describeProjectOptions{
				ProjectName:  "test-project",
				RevisionName: "test-revision",
				OutputFormat: "json",
			},
			revisionName: "test-revision",
			testServer: describeTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/revisions/test-revision/configuration" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"name": "test-project", "revision": "test-revision"}`))
					return true
				}
				return false
			}),
		},
		"valid project with version": {
			options: describeProjectOptions{
				ProjectName:  "test-project",
				VersionName:  "test-version",
				OutputFormat: "json",
			},
			testServer: describeTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/backend/projects/test-project/versions/test-version/configuration" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"name": "test-project", "revision": "test-version"}`))
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

			err = describeProject(ctx, client, testCase.options)

			if testCase.expectError {
				require.Error(t, err)
				require.EqualError(t, err, testCase.expectedErrorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func describeTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
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
