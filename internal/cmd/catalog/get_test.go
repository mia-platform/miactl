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

package catalog

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
	"github.com/mia-platform/miactl/internal/resources/catalog"
)

const (
	mockCompanyID       = "some-company-id"
	mockItemID          = "some-item-id"
	mockVersion         = "1.0.0"
	validBodyJSONString = `{
		"_id":"1234567890abcdefg",
		"name":"RocketScience 101: Hello Universe Example",
		"description":"A simple Hello Universe example based on Rocket-Launcher's Interstellar Template.",
		"type":"example",
		"itemId":"some-item-id",
		"version":{
		   "name":"1.0.0",
		   "releaseNote":"some release note"
		},
		"imageUrl":"/v2/files/download/rocket-launch-image.png",
		"supportedByImageUrl":"/v2/files/download/rocket-science-logo.png",
		"supportedBy":"NASA's Humor Department",
		"documentation":{
		   "type":"markdown",
		   "url":"https://raw.githubusercontent.com/rocket-launcher/Interstellar-Hello-Universe-Example/master/README.md"
		},
		"categoryId":"rocketScience",
		"resources":{
		   "services":{
			  "rocket-science-hello-universe-example":{
				 "archiveUrl":"https://github.com/rocket-launcher/Interstellar-Hello-Universe-Example/archive/master.tar.gz",
				 "containerPorts":[
					{
					   "name":"spaceport",
					   "from":80,
					   "to":3000,
					   "protocol":"TCP"
					}
				 ],
				 "type":"template",
				 "name":"rocket-science-hello-universe-example",
				 "pipelines":{
					"space-station-ci":{
					   "path":"/projects/space-station%2Fpipelines-templates/repository/files/console-pipeline%2Frocket-template.gitlab-ci.yml/raw"
					}
				 }
			  }
		   }
		}
	 }`
)

func TestGetResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := GetCmd(opts)
		require.NotNil(t, cmd)
	})

	t.Run("should not run command when Console version is lower than 14.0.0", func(t *testing.T) {
		server := httptest.NewServer(unexecutedCmdMockServer(t))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = mockCompanyID
		opts.Endpoint = server.URL

		cmd := GetCmd(opts)
		cmd.SetArgs([]string{"get", "--item-id", mockItemID, "--version", mockVersion})

		err := cmd.Execute()
		require.ErrorIs(t, err, catalog.ErrUnsupportedCompanyVersion)
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
				getItemByItemIDAndVersionEndpointTemplate, mockCompanyID, mockItemID, mockVersion),
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
		itemID       string
		version      string

		expectError         bool
		expectedCalledCount int
	}{
		"valid get response - json": {
			outputFormat:        encoding.JSON,
			statusCode:          http.StatusOK,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			itemID:    mockItemID,
			version:   mockVersion,
		},
		"valid get response - yaml": {
			statusCode:          http.StatusOK,
			outputFormat:        encoding.YAML,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			itemID:    mockItemID,
			version:   mockVersion,
		},
		"invalid body response": {
			statusCode:          http.StatusOK,
			expectError:         true,
			invalidResponse:     true,
			outputFormat:        encoding.JSON,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			itemID:    mockItemID,
			version:   mockVersion,
		},
		"resource not found": {
			statusCode:          http.StatusNotFound,
			expectError:         true,
			outputFormat:        encoding.JSON,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			itemID:    mockItemID,
			version:   mockVersion,
		},
		"internal server error": {
			statusCode:          http.StatusInternalServerError,
			outputFormat:        encoding.JSON,
			expectError:         true,
			expectedCalledCount: 1,

			companyID: mockCompanyID,
			itemID:    mockItemID,
			version:   mockVersion,
		},
		"should throw error and not call endpoint with missing company id": {
			statusCode:   http.StatusOK,
			outputFormat: encoding.JSON,

			expectError:         true,
			expectedCalledCount: 0,

			companyID: "",
			itemID:    mockItemID,
			version:   mockVersion,
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
				testCase.itemID,
				testCase.version,
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
