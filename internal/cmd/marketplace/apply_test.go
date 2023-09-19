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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPathsFromDir(t *testing.T) {
	t.Run("should read all files in dir retrieving paths", func(t *testing.T) {
		dirPath := "./testdata/withoutErrors"

		found, err := buildPathsListFromDir(dirPath)
		require.NoError(t, err)
		require.Contains(t, found, "testdata/withoutErrors/validItem1.json")
		require.Contains(t, found, "testdata/withoutErrors/validYaml.yaml")
		require.Contains(t, found, "testdata/withoutErrors/validYaml.yml")
		require.Len(t, found, 3)
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
		require.ErrorContains(t, err, "errors in file ./testdata/invalidJson1.json")
		require.Nil(t, found)
	})

	t.Run("should return error if file is not valid yaml", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidYaml.yaml",
		}

		found, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "errors in file ./testdata/invalidYaml.yaml")
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

	t.Run("should return error if a file has unknown extensions, but others are valid", func(t *testing.T) {
		filePaths := []string{
			"./testdata/someFileNotYamlNotJson.txt",
			"./testdata/validItem1.json",
			"./testdata/validYaml.yaml",
			"./testdata/validYaml.yml",
		}

		found, err := buildApplyRequest(filePaths)
		require.Error(t, err)
		require.Nil(t, found)
	})

	t.Run("should return error if resources array is empty", func(t *testing.T) {
		filePaths := []string{}

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

const mockTenantId = "some-tenant-id"

var mockURI = fmt.Sprintf(applyEndpoint, mockTenantId)

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

	t.Run("should return response when is 200 OK", func(t *testing.T) {
		mockResponse := &ApplyResponse{
			Done: true,
			Items: []ApplyResponseItem{
				{
					ItemID:   "some-id",
					Name:     mockResName,
					Done:     true,
					Inserted: true,
					Updated:  false,
				},
			},
		}
		server := applyMockServer(t, http.StatusOK, mockResponse)
		defer server.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyMarketplaceResource(
			context.Background(),
			client,
			mockTenantId,
			validReqMock,
		)

		require.NoError(t, err)
		require.Equal(t, found, mockResponse)
	})
	t.Run("should return err if response is a http error", func(t *testing.T) {
		mockResponse := map[string]interface{}{
			"message":    "You are not allowed to perform the request!",
			"statusCode": http.StatusForbidden,
			"error":      "Forbidden",
		}
		server := applyMockServer(t, http.StatusForbidden, mockResponse)
		defer server.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyMarketplaceResource(
			context.Background(),
			client,
			mockTenantId,
			validReqMock,
		)

		require.EqualError(t, err, "You are not allowed to perform the request!")
		require.Nil(t, found)
	})

}

func TestPrintApplyOutcome(t *testing.T) {
	t.Run("should match snapshot with both valid files and validation errors", func(t *testing.T) {
		mockOutcome := &ApplyResponse{
			Done: false,
			Items: []ApplyResponseItem{
				{
					ItemID:           "id1",
					Name:             "some name 1",
					Done:             true,
					Inserted:         false,
					Updated:          true,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
				{
					ItemID:           "id2",
					Name:             "some name 2",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
				{
					ItemID:           "id3",
					Name:             "some name 3",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
				{
					ItemID:   "id4",
					Name:     "some name 4",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []ApplyResponseItemValidationError{
						{
							Message: "some validation error",
						},
					},
				},
			},
		}
		snaps.MatchSnapshot(t, buildOutcomeSummaryAsTables(mockOutcome))
	})

	t.Run("should match snapshot with validation errors only", func(t *testing.T) {
		mockOutcome := &ApplyResponse{
			Done: false,
			Items: []ApplyResponseItem{
				{
					ItemID:   "id3",
					Name:     "some name 3",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []ApplyResponseItemValidationError{
						{
							Message: "some validation error",
						},
					},
				},
				{
					ItemID:   "id4",
					Name:     "some name 4",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []ApplyResponseItemValidationError{
						{
							Message: "some other validation error",
						},
					},
				},
			},
		}
		snaps.MatchSnapshot(t, buildOutcomeSummaryAsTables(mockOutcome))
	})

	t.Run("should match snapshot with valid files only", func(t *testing.T) {
		mockOutcome := &ApplyResponse{
			Done: false,
			Items: []ApplyResponseItem{
				{
					ItemID:           "id1",
					Name:             "some name 1",
					Done:             true,
					Inserted:         false,
					Updated:          true,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
				{
					ItemID:           "id2",
					Name:             "some name 2",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
				{
					ItemID:           "id3",
					Name:             "some name 3",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []ApplyResponseItemValidationError{},
				},
			},
		}
		snaps.MatchSnapshot(t, buildOutcomeSummaryAsTables(mockOutcome))
	})
}

func applyMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var isReqOk = assert.Equal(t, mockURI, r.RequestURI) && assert.Equal(t, http.MethodPost, r.Method)
		if !isReqOk {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(statusCode)
		resBytes, err := json.Marshal(mockResponse)
		require.NoError(t, err)
		w.Write(resBytes)
	}))
}
