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

package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
)

func TestCompareVersion(t *testing.T) {
	defaultMajorInt, err := strconv.Atoi(defaultMajor)
	require.NoError(t, err)
	defaultMinorInt, err := strconv.Atoi(defaultMinor)
	require.NoError(t, err)

	testCases := map[string]struct {
		major         int
		minor         int
		check         bool
		testServer    *httptest.Server
		expectedError string
	}{
		"missing api use default version - check true minor": {
			major: defaultMajorInt,
			minor: defaultMinorInt,
			check: true,
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
			}),
		},
		"missing api use default version - check true major": {
			major: defaultMajorInt - 1,
			minor: defaultMinorInt + 1,
			check: true,
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
			}),
		},
		"missing api use default version - check false major": {
			major: defaultMajorInt + 1,
			minor: defaultMinorInt - 1,
			check: false,
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
			}),
		},
		"missing api use default version - check false minor": {
			major: defaultMajorInt,
			minor: defaultMinorInt + 1,
			check: false,
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
			}),
		},
		"other error return error": {
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"statusCode": 500, "message": "error from server"}`))
			}),
			expectedError: "error from server",
		},
		"successful get return a valid version - check true minor": {
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.Write([]byte(`{"major": "5", "minor":"11"}`))
			}),
			major: 5,
			minor: 10,
			check: true,
		},
		"successful get return a valid version - check true major": {
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.Write([]byte(`{"major": "5", "minor":"11"}`))
			}),
			major: 4,
			minor: 12,
			check: true,
		},
		"successful get return a valid version - check false major": {
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.Write([]byte(`{"major": "5", "minor":"10"}`))
			}),
			major: 6,
			minor: 10,
			check: false,
		},
		"successful get return a valid version - check false minor": {
			testServer: versionTestServer(t, func(w http.ResponseWriter) {
				w.Write([]byte(`{"major": "5", "minor":"10"}`))
			}),
			major: 5,
			minor: 11,
			check: false,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			server := test.testServer
			defer server.Close()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)
			defer cancel()

			check, err := VersionCheck(ctx, client, test.major, test.minor)
			if len(test.expectedError) > 0 {
				assert.Error(t, err)
				assert.False(t, check)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.check, check)
		})
	}
}

func versionTestServer(t *testing.T, response func(w http.ResponseWriter)) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "request not expected")
		case r.URL.Path == "/api/version" && r.Method == http.MethodGet:
			response(w)
		}
	}))
}
