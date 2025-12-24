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

package itd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
)

const (
	mockCompanyID       = "some-company-id"
	mockName            = "plugin"
	validBodyJSONString = `{
    "apiVersion": "software-catalog.mia-platform.eu/v1",
    "kind": "item-type-definition",
    "metadata": {
        "namespace": {
            "scope": "tenant",
            "id": "some-company-id"
        },
        "name": "plugin",
        "visibility": {
            "scope": "tenant",
            "ids": [
                "some-company-id"
            ]
        },
        "displayName": "Plugin",
        "tags": [
            "prova"
        ],
        "maintainers": [
            {
                "name": "ok"
            }
        ],
        "publisher": {
            "name": "publisher-name"
        }
    },
    "spec": {
        "type": "plugin",
        "scope": "tenant",
        "validation": {
            "mechanism": "json-schema",
            "schema": {}
        }
    },
    "__v": 2
  }`
)

func TestGetResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := GetCmd(opts)
		require.NotNil(t, cmd)
	})

	t.Run("should not run command when Console version is lower than 14.1.0", func(t *testing.T) {
		server := httptest.NewServer(unexecutedCmdMockServer(t))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = mockCompanyID
		opts.Endpoint = server.URL

		cmd := GetCmd(opts)
		cmd.SetArgs([]string{"get", "--name", mockName})

		err := cmd.Execute()
		require.ErrorIs(t, err, itd.ErrUnsupportedCompanyVersion)
	})
}

func getItemByTupleMockServer(
	t *testing.T,
	validResponse bool,
	statusCode int,
	calledCount *int,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calledCount++
		assert.Equal(t,
			fmt.Sprintf(
				getItdEndpoint, mockCompanyID, mockName),
			r.RequestURI,
		)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(statusCode)
		if statusCode == http.StatusNotFound || statusCode == http.StatusInternalServerError {
			return
		}
		if validResponse {
			w.Write([]byte(validBodyJSONString))
			return
		}
		w.Write([]byte("invalid response"))
	}))
}

func TestGetItemEncodedByTuple(t *testing.T) {
	clientConfig := &client.Config{
		Transport: http.DefaultTransport,
	}

	testCases := map[string]struct {
		invalidResponse bool
		statusCode      int

		outputFormat string
		companyID    string
		name         string

		expectError         bool
		expectedCalledCount int
	}{
		"valid get response - json": {
			outputFormat:        encoding.JSON,
			statusCode:          http.StatusOK,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			name:      mockName,
		},
		"valid get response - yaml": {
			statusCode:          http.StatusOK,
			outputFormat:        encoding.YAML,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			name:      mockName,
		},
		"invalid body response": {
			statusCode:          http.StatusOK,
			expectError:         true,
			invalidResponse:     true,
			outputFormat:        encoding.JSON,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			name:      mockName,
		},
		"resource not found": {
			statusCode:          http.StatusNotFound,
			expectError:         true,
			outputFormat:        encoding.JSON,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			name:      mockName,
		},
		"internal server error": {
			statusCode:          http.StatusInternalServerError,
			outputFormat:        encoding.JSON,
			expectError:         true,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			name:      mockName,
		},
		"should throw error and not call endpoint with missing company id": {
			statusCode:   http.StatusOK,
			outputFormat: encoding.JSON,

			expectError:         true,
			expectedCalledCount: 0,

			companyID: "",
			name:      mockName,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			calledCount := new(int)
			*calledCount = 0
			server := getItemByTupleMockServer(
				t,
				!testCase.invalidResponse,
				testCase.statusCode,
				calledCount,
			)
			defer server.Close()
			clientConfig.Host = server.URL
			client, err := client.APIClientForConfig(clientConfig)
			require.NoError(t, err)
			found, err := getItemEncodedWithFormat(
				t.Context(),
				client,
				testCase.companyID,
				testCase.name,
				testCase.outputFormat,
			)

			require.Equal(
				t,
				testCase.expectedCalledCount,
				*calledCount,
				"unexpected number of calls to endpoint",
			)

			if testCase.expectError {
				require.Empty(t, found)
				require.Error(t, err)
			} else {
				if testCase.outputFormat == encoding.JSON {
					require.JSONEq(t, validBodyJSONString, found)
				} else {
					foundMap := map[string]interface{}{}
					err := yaml.Unmarshal([]byte(found), &foundMap)
					require.NoError(t, err)

					expectedMap := map[string]interface{}{}
					err = yaml.Unmarshal([]byte(found), &expectedMap)
					require.NoError(t, err)

					require.Equal(t, expectedMap, foundMap)
				}
				require.NoError(t, err)
			}
		})
	}
}
