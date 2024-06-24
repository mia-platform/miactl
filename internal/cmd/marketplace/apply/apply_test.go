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
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
		require.NotEmpty(t, found)
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

		require.ErrorContains(t, err, "file malformed ./testdata/invalidJson1.json")
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if file is not valid yaml", func(t *testing.T) {
		filePaths := []string{
			"./testdata/invalidYaml.yaml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorContains(t, err, "file malformed ./testdata/invalidYaml.yaml")
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

	t.Run("should return error if two resources have the same itemId - both without version", func(t *testing.T) {
		filePaths := []string{
			"./testdata/validYaml.yaml",
			"./testdata/validYaml.yml",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errDuplicatedResIdentifier)
		require.Nil(t, foundApplyReq)
		require.Nil(t, foundResNameToFilePath)
	})

	t.Run("should return error if two resources have the same itemId and same version name", func(t *testing.T) {
		filePaths := []string{
			"./testdata/validItemWithVersion.json",
			"./testdata/validItemWithSameVersion.json",
		}

		foundApplyReq, foundResNameToFilePath, err := buildApplyRequest(filePaths)
		require.ErrorIs(t, err, errDuplicatedResIdentifier)
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
			context.TODO(),
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
			context.TODO(),
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
		require.Contains(t, found, "OBJECT ID")
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
		require.Contains(t, found, "OBJECT ID")
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
		require.Contains(t, found, "OBJECT ID")
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
			func(resources []interface{}) {
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
			},
			nil,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.TODO(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.NoError(t, err)
		require.Contains(t, found, "item-id-1")
		require.Contains(t, found, "id1")
	})

	t.Run("should upload images correctly - from items with same ids and different versions", func(t *testing.T) {
		mockUploadImageStatusCode := http.StatusOK
		mockApplyItemStatusCode := http.StatusOK

		mockPaths := []string{
			"./testdata/validItemWithImage.json",
			"./testdata/validItemWithVersion.json",
		}
		applyMockResponse := &marketplace.ApplyResponse{
			Done: true,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:               "id1",
					ItemID:           "miactl-test-with-image-and-local-path",
					Done:             true,
					Inserted:         true,
					Updated:          false,
					ValidationErrors: []marketplace.ApplyResponseItemValidationError{},
				},
				{
					ID:               "id1",
					ItemID:           "miactl-test-with-image-and-local-path",
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

		uploadImageCallIdx := 0
		mockServer := applyIntegrationMockServer(t,
			mockUploadImageStatusCode,
			mockApplyItemStatusCode,
			uploadMockResponse,
			applyMockResponse,
			func(resources []interface{}) {
				assertImageKeyIsReplacedWithImageURL(
					t,
					resources[0].(map[string]interface{}),
					imageKey,
					imageURLKey,
				)
				version, ok := resources[0].(map[string]interface{})["version"]
				assert.False(t, ok)
				assert.Nil(t, version)

				assertImageKeyIsReplacedWithImageURL(
					t,
					resources[1].(map[string]interface{}),
					imageKey,
					imageURLKey,
				)

				version, ok = resources[1].(map[string]interface{})["version"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotNil(t, version)

				versionName, ok := version.(map[string]interface{})["name"].(string)
				assert.True(t, ok)
				assert.Equal(t, "1.0.0", versionName)
			},
			func(mf *multipart.Form) {
				t.Helper()

				require.LessOrEqual(t, uploadImageCallIdx, 1, "too many calls to upload image endpoint")
				if uploadImageCallIdx == 0 {
					require.Equal(t, "miactl-test-with-image-and-local-path", mf.Value["itemId"][0])
					require.Equal(t, imageAssetType, mf.Value["assetType"][0])
					require.Equal(t, mockTenantID, mf.Value["tenantId"][0])
					require.Nil(t, mf.Value["version"])

					require.Equal(t, "imageTest.png", mf.File[multipartFieldName][0].Filename)
				}
				if uploadImageCallIdx == 1 {
					require.Equal(t, "miactl-test-with-image-and-local-path", mf.Value["itemId"][0])
					require.Equal(t, imageAssetType, mf.Value["assetType"][0])
					require.Equal(t, mockTenantID, mf.Value["tenantId"][0])
					require.Equal(t, "1.0.0", mf.Value["version"][0])

					require.Equal(t, "imageTest2.png", mf.File[multipartFieldName][0].Filename)
				}

				uploadImageCallIdx++
			},
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.TODO(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.NoError(t, err)
		require.Contains(t, found, "miactl-test-with-image-and-local-path")
		require.Contains(t, found, "id1")
	})

	t.Run("should return the error message returned from server if image upload fails", func(t *testing.T) {
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
			nil,
			nil,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.TODO(),
			client,
			mockTenantID,
			mockPaths,
		)
		require.ErrorContains(t, err, mockErrorMsg)
		require.Zero(t, found)
	})

	t.Run("should return the error message returned from server if apply item fails", func(t *testing.T) {
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
			func(resources []interface{}) {
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
			},
			nil,
		)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := applyItemsFromPaths(
			context.TODO(),
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
	assertResourcesFn func(resources []interface{}),
	assertImageUploadBody func(mf *multipart.Form),
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

			if assertResourcesFn != nil {
				assertResourcesFn(resources)
			}

			applyRequestHandler(t, w, r, applyItemStatusCode, applyMockResponse)
		case fmt.Sprintf(uploadImageEndpointTemplate, mockTenantID):
			uploadImageHandler(t, w, r, uploadImageStatusCode, uploadMockResponse)
			if assertImageUploadBody != nil {
				err := r.ParseMultipartForm(5 * 10000)
				require.NoError(t, err)
				mf := r.MultipartForm
				require.NotNil(t, mf.Value)
				assertImageUploadBody(mf)
			}

		default:
			require.FailNowf(t, "invalid request URI", "invalid request URI: %s", r.RequestURI)
		}
	}))
}

func TestProcessItemImages(t *testing.T) {
	testCases := []struct {
		testName string

		itemJSON                 string
		itemVersionToFilePathMap map[string]string

		expectedErr  error
		expectedItem *marketplace.Item
	}{
		{
			testName: "should upload file correctly - without version",

			itemJSON: `{
				"itemId": "some-id",
				"image": {
					"localPath": "./testdata/imageTest.png"
				}
			}`,
			expectedItem: &marketplace.Item{
				"itemId":    "some-id",
				imageURLKey: mockImageURLLocation,
			},
			itemVersionToFilePathMap: map[string]string{
				"itemId": "./testdata/imageTest.png",
			},
		},
		{
			testName: "should upload file correctly - with version",

			itemJSON: `{
				"itemId": "some-id",
				"image": {
					"localPath": "./testdata/imageTest.png"
				},
				"version": {
					"name": "1.0.0",
					"releaseNotes": "some release note"
				}
			}`,
			expectedItem: &marketplace.Item{
				"itemId":    "some-id",
				imageURLKey: mockImageURLLocation,
				"version": marketplace.Item{
					"name":         "1.0.0",
					"releaseNotes": "some release note",
				},
			},
			itemVersionToFilePathMap: map[string]string{
				"itemId": "./testdata/imageTest.png",
			},
		},
	}

	for _, testCase := range testCases {
		applyMockResponse := &marketplace.ApplyResponse{
			Done: true,
			Items: []marketplace.ApplyResponseItem{
				{
					ID:               "id1",
					ItemID:           "some-id",
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

		server := applyIntegrationMockServer(t,
			http.StatusOK,
			http.StatusOK,
			uploadMockResponse,
			applyMockResponse,
			nil,
			nil,
		)
		defer server.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		mockItem := &marketplace.Item{}
		err = yaml.Unmarshal([]byte(testCase.itemJSON), mockItem)
		require.NoError(t, err)

		err = processItemImages(
			context.TODO(),
			client,
			mockTenantID,
			mockItem,
			testCase.itemVersionToFilePathMap,
		)

		if testCase.expectedErr != nil {
			require.ErrorIs(t, err, testCase.expectedErr)
			continue
		}

		require.NoError(t, err)

		// the item is changed in-place, so they should be equal
		require.Equal(t, testCase.expectedItem, mockItem)
	}
}

func TestBuildIdentifier(t *testing.T) {
	testCases := []struct {
		testName    string
		itemJSON    string
		expected    string
		expectedErr error
	}{
		{
			"should replace only itemID when version is not provided",
			`{
				"itemId": "some-id"
			}`,
			"some-id",
			nil,
		},
		{
			"should return error if itemID is not a string",
			`{
				"itemId": 5
			}`,
			"",
			errResItemIDNotAString,
		},
		{
			"should return error if name is not defined",
			`{
				"itemId":  "some-other-id",
				"version": {}
			}`,
			"",
			marketplace.ErrVersionNameNotAString,
		},
		{
			"should return error if name is not a string",
			`{
				"itemId": "some-id",
				"version": {
					"name": 1
				}
			}`,
			"",
			marketplace.ErrVersionNameNotAString,
		},
		{
			"should return itemID concatenated with version name when version name is present",
			`{
				"itemId": "some-id",
				"version": {
					"name": "1.0.0"
				}
			}`,
			"some-id1.0.0",
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			parsedItem := &marketplace.Item{}
			err := yaml.Unmarshal([]byte(tt.itemJSON), parsedItem)
			require.NoError(t, err)

			found, err := buildItemIdentifier(parsedItem)
			if tt.expectedErr != nil {
				require.Zero(t, found)
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, tt.expected, found)
		})
	}
}

func applyMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyRequestHandler(t, w, r, statusCode, mockResponse)
	}))
}
