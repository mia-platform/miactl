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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockResourceID = "resource-id"
)

func TestGetResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := GetCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestGetResourceById(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		err          bool
		outputFormat string
	}{
		"valid get response - json": {
			server: getByIDMockServer(t, true, http.StatusOK),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			outputFormat: JSON,
		},
		"valid get response - yaml": {
			server: getByIDMockServer(t, true, http.StatusOK),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			outputFormat: YAML,
		},
		"invalid body response": {
			server: getByIDMockServer(t, false, http.StatusOK),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err:          true,
			outputFormat: JSON,
		},
		"resource not found": {
			server: getByIDMockServer(t, true, http.StatusNotFound),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err:          true,
			outputFormat: JSON,
		},
		"internal server error": {
			server: getByIDMockServer(t, true, http.StatusInternalServerError),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err:          true,
			outputFormat: JSON,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			err = getMarketplaceResource(client, mockResourceID, testCase.outputFormat)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getByIDMockServer(t *testing.T, validResponse bool, statusCode int) *httptest.Server {
	t.Helper()
	validBodyString := `{
		"_id": "1234567890abcdefg",
		"name": "RocketScience 101: Hello Universe Example",
		"description": "A simple Hello Universe example based on Rocket-Launcher's Interstellar Template.",
		"type": "example",
		"imageUrl": "/v2/files/download/rocket-launch-image.png",
		"supportedByImageUrl": "/v2/files/download/rocket-science-logo.png",
		"supportedBy": "NASA's Humor Department",
		"documentation": {
			"type": "markdown",
			"url": "https://raw.githubusercontent.com/rocket-launcher/Interstellar-Hello-Universe-Example/master/README.md"
		},
		"categoryId": "rocketScience",
		"resources": {
			"services": {
				"rocket-science-hello-universe-example": {
					"archiveUrl": "https://github.com/rocket-launcher/Interstellar-Hello-Universe-Example/archive/master.tar.gz",
					"containerPorts": [
						{
							"name": "spaceport",
							"from": 80,
							"to": 3000,
							"protocol": "TCP"
						}
					],
					"type": "template",
					"name": "rocket-science-hello-universe-example",
					"pipelines": {
						"space-station-ci": {
							"path": "/projects/space-station%2Fpipelines-templates/repository/files/console-pipeline%2Frocket-template.gitlab-ci.yml/raw"
						}
					}
				}
			}
		}
	}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != fmt.Sprintf(getMarketplaceEndpoint, mockResourceID) && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(statusCode)
		if statusCode == http.StatusNotFound || statusCode == http.StatusInternalServerError {
			return
		}
		if validResponse {
			w.Write([]byte(validBodyString))
			return
		}
		w.Write([]byte("invalid json"))
	}))
}
