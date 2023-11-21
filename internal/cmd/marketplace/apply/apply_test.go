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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/require"
)

func TestApplyBuildPathsFromDir(t *testing.T) {
	t.Run("should read all files in dir retrieving paths", func(t *testing.T) {
		dirPath := "./testdata/subdir"

		found, err := buildFilePathsList([]string{dirPath})
		require.NoError(t, err)
		require.Contains(t, found, "testdata/subdir/validItem1.json")
		require.Contains(t, found, "testdata/subdir/validYaml.yaml")
		require.Contains(t, found, "testdata/subdir/validYaml.yml")
		require.Contains(t, found, "testdata/subdir/validItemWithImage.json")
		require.Len(t, found, 4)
	})

	t.Run("should NOT return error due to file with bad extension", func(t *testing.T) {
		dirPath := "./testdata/"

		found, err := buildFilePathsList([]string{dirPath})
		require.Equal(t, []string{"testdata/invalidJson1.json", "testdata/invalidYaml.yaml", "testdata/invalidYml.yml", "testdata/subdir/validItem1.json", "testdata/subdir/validItemWithImage.json", "testdata/subdir/validYaml.yaml", "testdata/subdir/validYaml.yml", "testdata/validItem1.json", "testdata/validItemWithImage.json", "testdata/validItemWithImage2.json", "testdata/validYaml.yaml", "testdata/validYaml.yml", "testdata/yamlWithImage.yml"}, found)
		require.NoError(t, err)
	})
}

func TestApplyBuildApplyRequest(t *testing.T) {
	t.Run("should read file contents parsing them to json", func(t *testing.T) {
		filePaths := []string{
			"./testdata/validItem1.json",
			"./testdata/validYaml.yaml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.NoError(t, err)
		require.NotNil(t, foundApplyReq)
		require.NotEmpty(t, foundApplyReq.Resources)
		require.Equal(t, foundResNameToFilePath, map[string]string{
			"miactl-test-json": "./testdata/validItem1.json",
			"miactl-test":      "./testdata/validYaml.yaml",
		})
	})

	t.Run("should return error if file is not valid json", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidJson1.json",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)

		require.ErrorContains(t, err, "errors in file ./testdata/invalidJson1.json")
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if file is not valid yaml", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidYaml.yaml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "errors in file ./testdata/invalidYaml.yaml")
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if file is not found", func(t *testing.T) {
		filePaths := []string{
			"./I/do/not/exist.json",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.Error(t, err)
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should ignore a file it it has unknown extensions, and should parse only valid files", func(t *testing.T) {
		filePaths := []string{
			"./testdata/someFileNotYamlNotJson.txt",
			"./testdata/validItem1.json",
			"./testdata/validYaml.yaml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.NoError(t, err)
		require.NotNil(t, foundApplyReq)
		require.Len(t, foundApplyReq.Resources, 2)
		require.Equal(t, foundResNameToFilePath, map[string]string{
			"miactl-test-json": "./testdata/validItem1.json",
			"miactl-test":      "./testdata/validYaml.yaml",
		})
	})

	t.Run("should return error if resources array is empty", func(t *testing.T) {
		filePaths := []string{}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errNoValidFilesProvided)
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if resources are all with invalid file extension", func(t *testing.T) {
		filePaths := []string{
			"./testdata/someFileNotYamlNotJson.txt",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errNoValidFilesProvided)
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if two resources have the same itemId", func(t *testing.T) {
		filePaths := []string{
			"./testdata/validYaml.yaml",
			"./testdata/validYaml.yml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errDuplicatedResItemID)
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})
}

const mockTenantID = "some-tenant-id"

var mockURI = fmt.Sprintf(applyEndpointTemplate, mockTenantID)

func TestApplyApplyResourceCmd(t *testing.T) {
	mockItemID := "some-item-id"
	validReqMock := &marketplace.ApplyRequest{
		Resources: []*marketplace.Item{
			{
				"categoryId":    "devportal",
				"imageUrl":      "some/path/to/image.png",
				"name":          "some name",
				"itemId":        mockItemID,
				"releaseStage":  "",
				"repositoryUrl": "https://example.com/repo",
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
							"dockerImage":   "containers.example.com/some-image:latest",
							"name":          "api-portal",
							"repositoryUrl": "https://example.com/repo",
							"type":          "plugin",
						},
					},
				},
				"supportedByImageUrl": "/some/path/to/image.png",
				"tenantId":            "team-rocket-test",
				"type":                "plugin",
			},
		},
	}

	t.Run("should return response when is 200 OK", func(t *testing.T) {
		mockResponse := &marketplace.ApplyResponse{
			Done: true,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:       "id1",
					ItemID:   "some-item-id",
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
			mockTenantID,
			validReqMock,
		)

		require.NoError(t, err)
		require.Equal(t, found, mockResponse)
	})
	t.Run("should return err if response is a http error", func(t *testing.T) {
		mockErrorResponse := map[string]interface{}{
			"message":    "You are not allowed to perform the request!",
			"statusCode": http.StatusForbidden,
			"error":      "Forbidden",
		}
		server := applyMockServer(t, http.StatusForbidden, mockErrorResponse)
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
			mockTenantID,
			validReqMock,
		)

		require.EqualError(t, err, "You are not allowed to perform the request!")
		require.Nil(t, found)
	})
}

func TestApplyPrintApplyOutcome(t *testing.T) {
	t.Run("should contain both valid files and validation errors", func(t *testing.T) {
		mockOutcome := &marketplace.ApplyResponse{
			Done: false,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:               "id1",
					ItemID:           "item-id-1",
					Done:             true,
					Inserted:         false,
					Updated:          true,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:               "id2",
					ItemID:           "item-id-2",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:               "id3",
					ItemID:           "item-id-3",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:       "id4",
					ItemID:   "item-id-4",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{
						{
							Message: "some validation error",
						},
					},
				},
			},
		}
		found := buildOutcomeSummaryAsTables(mockOutcome)
		require.Contains(t, found, "3 of 4 items have been successfully applied:")
		require.Contains(t, found, "id1")
		require.Contains(t, found, "item-id-1")
		require.Contains(t, found, "id2")
		require.Contains(t, found, "item-id-2")
		require.Contains(t, found, "id3")
		require.Contains(t, found, "item-id-3")
		require.Contains(t, found, "1 of 4 items have not been applied due to validation errors:")
		require.Contains(t, found, "id4")
		require.Contains(t, found, "item-id-4")
		require.Contains(t, found, "some validation error")
		require.Contains(t, found, "ID")
		require.Contains(t, found, "ITEM ID")
	})

	t.Run("should show validation errors only when input does not contain successful applies", func(t *testing.T) {
		mockOutcome := &marketplace.ApplyResponse{
			Done: false,
			Items: []marketplace.ApplyResponseItem{
				{
					ItemID:   "some-item-id-1",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{
						{
							Message: "some validation error",
						},
					},
				},
				{
					ItemID:   "some-item-id-2",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{
						{
							Message: "some other validation error",
						},
						{
							Message: "and also some other validation error",
						},
					},
				},
				{
					ID:       "id3",
					ItemID:   "some-item-id-3",
					Done:     false,
					Inserted: false,
					Updated:  false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{
						{
							Message: "some other very very long validation error",
						},
					},
				},
			},
		}
		found := buildOutcomeSummaryAsTables(mockOutcome)
		require.NotContains(t, found, "items have been successfully applied:")
		require.Contains(t, found, "3 of 3 items have not been applied due to validation errors:")
		require.Contains(t, found, "some validation error")
		require.Contains(t, found, "some other validation error")
		require.Contains(t, found, "some other very very long validation error")
		require.Contains(t, found, "N/A")
		require.Contains(t, found, "id3")
		require.Contains(t, found, "some-item-id-3")
		require.Contains(t, found, "some-item-id-2")
		require.Contains(t, found, "some-item-id-1")
		require.Contains(t, found, "ID")
		require.Contains(t, found, "ITEM ID")
	})

	t.Run("should match with valid files only", func(t *testing.T) {
		mockOutcome := &marketplace.ApplyResponse{
			Done: false,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:               "id1",
					ItemID:           "some-item-id-1",
					Done:             true,
					Inserted:         false,
					Updated:          true,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:               "id2",
					ItemID:           "some-item-id-2",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:               "id3",
					ItemID:           "some-item-id-3",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
			},
		}
		found := buildOutcomeSummaryAsTables(mockOutcome)
		require.Contains(t, found, "3 of 3 items have been successfully applied:")
		require.Contains(t, found, "id1")
		require.Contains(t, found, "some-item-id-1")
		require.Contains(t, found, "id2")
		require.Contains(t, found, "some-item-id-2")
		require.Contains(t, found, "id3")
		require.Contains(t, found, "some-item-id-3")
		require.Contains(t, found, "ID")
		require.Contains(t, found, "ITEM ID")
		require.NotContains(t, found, "items have not been applied due to validation errors:")
	})
}

const mockImageURLLocation = "some/fancy/location"

func TestApplyIntegration(t *testing.T) {
	t.Run("should upload images correctly", func(t *testing.T) {
		mockUploadImageStatusCode := http.StatusOK
		mockApplyItemStatusCode := http.StatusOK

		mockPaths := []string{
			"./testdata/validItemWithImage.json",
			"./testdata/validItemWithImage2.json",
			"./testdata/yamlWithImage.yml",
			"./testdata/subdir/validItemWithImage.json",
		}
		applyMockResponse := &marketplace.ApplyResponse{
			Done: true,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:               "id1",
					ItemID:           "item-id-1",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
			},
		}
		uploadMockResponse := &marketplace.UploadImageResponse{
			Location: mockImageURLLocation,
		}

		mockServer := applyIntegrationMockServer(t,
			mockUploadImageStatusCode,
			mockApplyItemStatusCode,
			uploadMockResponse,
			applyMockResponse,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.Background(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.NoError(t, err)
		require.Contains(t, found, "item-id-1")
		require.Contains(t, found, "id1")
	})

	t.Run("should return the correct error message if image upload fails", func(t *testing.T) {
		mockUploadImageStatusCode := http.StatusBadRequest
		mockApplyItemStatusCode := http.StatusOK

		mockPaths := []string{
			"./testdata/validItemWithImage.json",
			"./testdata/validItemWithImage2.json",
			"./testdata/yamlWithImage.yml",
			"./testdata/subdir/validItemWithImage.json",
		}
		applyMockResponse := &marketplace.ApplyResponse{
			Done: true,
			Items: []marketplace.ApplyResponseItem{
				{
					ItemID:           "item-id-1",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
			},
		}
		mockErrorMsg := "upload image: some mock error message"
		uploadMockErrorResponse := map[string]interface{}{
			"error":      "upload image: some mock error",
			"message":    mockErrorMsg,
			"statusCode": mockUploadImageStatusCode,
		}
		mockServer := applyIntegrationMockServer(t,
			mockUploadImageStatusCode,
			mockApplyItemStatusCode,
			uploadMockErrorResponse,
			applyMockResponse,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.Background(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.ErrorContains(t, err, mockErrorMsg)
		require.Zero(t, found)
	})

	t.Run("should return the correct error message if apply item fails", func(t *testing.T) {
		mockUploadImageStatusCode := http.StatusOK
		mockApplyItemStatusCode := http.StatusBadRequest

		mockPaths := []string{
			"./testdata/validItemWithImage.json",
			"./testdata/validItemWithImage2.json",
			"./testdata/yamlWithImage.yml",
			"./testdata/subdir/validItemWithImage.json",
		}
		mockErrorMsg := "apply item: some mock error message"
		applyMockErrorResponse := map[string]interface{}{
			"error":      "apply item: some mock error",
			"message":    mockErrorMsg,
			"statusCode": mockUploadImageStatusCode,
		}
		uploadMockResponse := &marketplace.UploadImageResponse{
			Location: mockImageURLLocation,
		}

		mockServer := applyIntegrationMockServer(t,
			mockUploadImageStatusCode,
			mockApplyItemStatusCode,
			uploadMockResponse,
			applyMockErrorResponse,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.Background(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.ErrorContains(t, err, mockErrorMsg)
		require.Zero(t, found)
	})
}

func TestApplyConcatPathIfRelative(t *testing.T) {
	t.Run("should concat relative paths", func(t *testing.T) {
		found := concatPathDirToFilePathIfRelative("/some/path/to/file.json", "./image.png")
		require.Equal(t, "/some/path/to/image.png", found)
	})
	t.Run("should return an abs path", func(t *testing.T) {
		found := concatPathDirToFilePathIfRelative("/some/path/to/file.json", "/some/absolute/path/image.png")
		require.Equal(t, "/some/absolute/path/image.png", found)
	})
}

func applyRequestHandler(t *testing.T, w http.ResponseWriter, r *http.Request, statusCode int, mockResponse interface{}) {
	t.Helper()
	require.Equal(t, mockURI, r.RequestURI)
	require.Equal(t, http.MethodPost, r.Method)

	w.WriteHeader(statusCode)
	resBytes, err := json.Marshal(mockResponse)
	require.NoError(t, err)
	w.Write(resBytes)
}

func assertImageKeyIsReplacedWithImageURL(t *testing.T, resource map[string]interface{}, objKey, urlKey string) {
	t.Helper()

	require.NotContains(t, resource, objKey)
	require.Contains(t, resource, urlKey)
	require.Equal(t, mockImageURLLocation, resource[urlKey].(string))
}

func applyIntegrationMockServer(
	t *testing.T,
	uploadImageStatusCode, applyItemStatusCode int,
	uploadMockResponse, applyMockResponse interface{},
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case fmt.Sprintf(applyEndpointTemplate, mockTenantID):
			if applyItemStatusCode != http.StatusOK {
				applyRequestHandler(t, w, r, applyItemStatusCode, applyMockResponse)
				break
			}
			foundBodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			foundBody := make(map[string]interface{})
			err = json.Unmarshal(foundBodyBytes, &foundBody)
			require.NoError(t, err)
			require.Contains(t, foundBody, "resources")
			resources := foundBody["resources"].([]interface{})

			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[0].(map[string]interface{}),
				imageKey,
				imageURLKey,
			)
			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[1].(map[string]interface{}),
				imageKey,
				imageURLKey,
			)
			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[1].(map[string]interface{}),
				supportedByImageKey,
				supportedByImageURLKey,
			)
			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[2].(map[string]interface{}),
				imageKey,
				imageURLKey,
			)
			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[2].(map[string]interface{}),
				supportedByImageKey,
				supportedByImageURLKey,
			)
			assertImageKeyIsReplacedWithImageURL(
				t,
				resources[3].(map[string]interface{}),
				imageKey,
				imageURLKey,
			)

			applyRequestHandler(t, w, r, applyItemStatusCode, applyMockResponse)
		case fmt.Sprintf(uploadImageEndpointTemplate, mockTenantID):
			uploadImageHandler(t, w, r, uploadImageStatusCode, uploadMockResponse)
		default:
			require.FailNowf(t, "invalid request URI", "invalid request URI: %s", r.RequestURI)
		}
	}))
}

func applyMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyRequestHandler(t, w, r, statusCode, mockResponse)
	}))
}
