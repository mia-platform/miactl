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
		require.Contains(t, found, "invalidJson1.json")
		require.Contains(t, found, "invalidYaml.yaml")
		require.Contains(t, found, "invalidYml.yml")
		require.Contains(t, found, "validItem1.json")
		require.NotContains(t, found, "someFileNotYamlNotJson.txt")
		require.Equal(t, 4, len(found))
	})
}

func TestBuildResourcesList(t *testing.T){
	t.Run("should read file contents parsing them to json", func(t *testing.T) {
		dirPath := "./testdata"
		found, err := buildPathsListFromDir(dirPath)
		require.NoError(t, err)

		resources, err := buildResourcesList(found)

		require.NoError(t, err)
		require.NotEmpty(t, resources)
		
	})

	t.Run("should return error if file is not valid json", func(t *testing.T) {

	})

	t.Run("should return error if file is not valid yaml", func(t *testing.T) {

	})
}