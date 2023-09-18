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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestValidateResource(t *testing.T) {
	t.Run("should return error if resource does not contain a name", func(t *testing.T) {
		mockResource := &MarketplaceResource{
			"foo": "bar",
		}

		err := validateResource(mockResource)
		require.EqualError(t, err, `required field "name" was not found in the resource`)
	})

	t.Run("should not return error if resource contains a name", func(t *testing.T) {
		mockResource := &MarketplaceResource{
			"foo":  "bar",
			"name": "some name",
		}

		err := validateResource(mockResource)
		require.NoError(t, err)
	})
}

func TestApplyResourceCmd(t *testing.T) {
	mockResName := "API Portal by miactl test"
	validReqMock := &ApplyRequest{
		Resources: []*MarketplaceResource{
			{
				"_id":         "6504773582a6722338be0e25",
				"categoryId":  "devportal",
				"description": "Use Mia-Platform core API Portal to expose the swagger documentation of your development services in one unique place.",
				"documentation": map[string]interface{}{
					"type": "externalLink",
					"url":  "https://docs.mia-platform.eu/docs/runtime_suite/api-portal/overview",
				},
				"imageUrl":      "/v2/files/download/e83a1e48-fca7-4114-a244-1a69c0c4e7b2.png",
				"name":          mockResName,
				"releaseStage":  "",
				"repositoryUrl": "https://git.tools.mia-platform.eu/platform/api-portal/website",
				"resources": map[string]interface{}{
					"services": map[string]interface{}{
						"api-portal": map[string]interface{}{
							"componentId": "api-portal",
							"containerPorts": []map[string]interface{}{
								{
									"from":     80,
									"name":     "http",
									"protocol": "TCP",
									"to":       8080,
								},
							},
							"defaultEnvironmentVariables": []map[string]interface{}{
								{
									"name":      "HTTP_PORT",
									"value":     "8080",
									"valueType": "plain",
								},
								{
									"name":      "ANTI_ZOMBIE_PORT",
									"value":     "090909",
									"valueType": "plain",
								},
							},
							"defaultLogParser": "mia-nginx",
							"defaultProbes": map[string]interface{}{
								"liveness": map[string]interface{}{
									"path": "/index.html",
								},
								"readiness": map[string]interface{}{
									"path": "/index.html",
								},
							},
							"defaultResources": map[string]interface{}{
								"memoryLimits": map[string]interface{}{
									"max": "25Mi",
									"min": "5Mi",
								},
							},
							"description":   "Use Mia-Platform core API Portal to expose the swagger documentation of your development services in one unique place.",
							"dockerImage":   "nexus.mia-platform.eu/api-portal/website:1.16.6",
							"name":          "api-portal",
							"repositoryUrl": "https://git.tools.mia-platform.eu/platform/api-portal/website",
							"type":          "plugin",
						},
					},
				},
				"supportedByImageUrl": "/v2/files/download/83b11dd9-41f6-4920-bb2d-2a809e944851.png",
				"tenantId":            "team-rocket-test",
				"type":                "plugin",
			},
		},
	}
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companiesURI string
		err          bool
	}{
		"valid apply response": {
			server: applyMockServer(
				t,
				http.StatusOK,
				&ApplyResponse{
					Done: true,
					Items: []ApplyResponseItem{
						{
							ItemID: "some-id",
							Name:   mockResName,

							Done:     true,
							Inserted: true,
							Updated:  false,
						},
					},
				},
			),
			companiesURI: applyEndpoint,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)

			found, err := applyMarketplaceResource(client, "some-id", validReqMock)
			require.NoError(t, err)

			require.Contains(t, found, mockResName)

			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func applyMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var isReqOk = assert.Equal(t, applyEndpoint, r.RequestURI) && assert.Equal(t, http.MethodPost, r.Method)
		if !isReqOk {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(statusCode)
		res := []byte(`{"name":"ciao"}`)
		err := json.Unmarshal(res, mockResponse)
		require.NoError(t, err)
		w.Write(res)
	}))
}
