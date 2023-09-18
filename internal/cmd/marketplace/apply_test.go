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

package marketplace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func applyMockServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var isReqOk = assert.Equal(t, applyEndpoint, r.RequestURI) && assert.Equal(t, http.MethodPost, r.Method)
		if !isReqOk {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(statusCode)
	}))
}

func TestApplyResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		t.Skip()
	})
}

func TestBuildPathsFromDir(t *testing.T) {
	t.Run("should read all files in dir, ignoring non json and non yaml files, retrieving paths", func(t *testing.T) {
		dirPath := "./testdata"

		found, err := buildPathsListFromDir(dirPath)
		require.NoError(t, err)
		require.Contains(t, found, "testdata/invalidJson1.json")
		require.Contains(t, found, "testdata/invalidYaml.yaml")
		require.Contains(t, found, "testdata/invalidYml.yml")
		require.Contains(t, found, "testdata/validItem1.json")
		require.NotContains(t, found, "testdata/someFileNotYamlNotJson.txt")
		require.Len(t, found, 6)
	})
}

func TestBuildResourcesList(t *testing.T) {
	t.Run("should read file contents parsing them to json", func(t *testing.T) {
		filePaths := []string{
			"./testdata/validItem1.json",
			"./testdata/validYaml.yaml",
			"./testdata/validYaml.yml",
		}

		found, err := buildApplyRequest(filePaths)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.NotEmpty(t, found.Resources)
	})

	t.Run("should return error if file is not valid json", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidJson1.json",
		}

		found, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "error parsing file")
		require.Nil(t, found)
	})

	t.Run("should return error if file is not valid yaml", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidYaml.yaml",
		}

		found, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "error parsing file")
		require.Nil(t, found)
	})

	t.Run("should return error if file is not found", func(t *testing.T) {
		filePaths := []string{
			"./I/do/not/exist.json",
		}

		found, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "error opening file")
		require.Nil(t, found)
	})

	t.Run("should not return error if a file has unknown extensions, but others are valid", func(t *testing.T) {
		filePaths := []string{
			"./testdata/someFileNotYamlNotJson.txt",
			"./testdata/validItem1.json",
			"./testdata/validYaml.yaml",
			"./testdata/validYaml.yml",
		}

		found, err := buildApplyRequest(filePaths)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.NotEmpty(t, found.Resources)
		require.Len(t, found.Resources, 3)
	})

	t.Run("should return error if resources array is empty, i.e. only files with bad extensions as input", func(t *testing.T) {
		filePaths := []string{
			"./testdata/someFileNotYamlNotJson.txt",
		}

		found, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errNoValidFilesProvided)
		require.Nil(t, found)
	})
}
