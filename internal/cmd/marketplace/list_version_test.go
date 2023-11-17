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

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	listVersionsMockResponseBody = `[
    {
		"name": "Some Awesome Service",
		"description": "The Awesome Service allows to do some amazing stuff.",
		"version": "1.0.0",
		"reference": "655342ce0f991db238fd73e4",
		"security": false,
		"releaseNote": "-",
		"visibility": {
		  "public": true
		}
	  }
]`
)

func TestNewListVersionsCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListVersionCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestBuildMarketplaceItemVersionList(t *testing.T) {
	testCases := map[string]struct {
		releases         []marketplace.Release
		expectedContains []string
	}{
		"valid get response": {
			releases: []marketplace.Release{
				{
					Version:     "1.0.0",
					Name:        "Some Awesome Service",
					Description: "The Awesome Service allows to do some amazing stuff.",
				},
			},
			expectedContains: []string{
				"VERSION", "NAME", "DESCRIPTION",
				"1.0.0", "Some Awesome Service", "The Awesome Service allows",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			found, err := buildItemVersionList(testCase.releases)
			assert.NoError(t, err)
			assert.NotZero(t, found)
			for _, expected := range testCase.expectedContains {
				assert.Contains(t, found, expected)
			}
		})
	}
}

func mockListVersionsServer(t *testing.T, validResponse bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != listMarketplaceEndpoint && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(http.StatusOK)
		if validResponse {
			w.Write([]byte(mockResponseBody))
			return
		}
		w.Write([]byte(`{"message": "invalid json"}`))
	}))
}
