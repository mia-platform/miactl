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
	"os"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/require"
)

func TestApplyValidateImageURLs(t *testing.T) {
	t.Run("should throw error with an item that contains both image and imageURL", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]interface{}{
				localPathKey: "some/local/path/image.jpg",
			},
			imageURLKey: "http://some.url",
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageURLConflict)
		require.Zero(t, found)
	})

	t.Run("should return local path if element contains image", func(t *testing.T) {
		mockItemJson := []byte(`{
			"image": {
				"localPath": "some/local/path/image.jpg"
			}
		}`)
		mockItem := &marketplace.Item{}
		err := json.Unmarshal(mockItemJson, mockItem)
		require.NoError(t, err)

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Equal(t, found, "some/local/path/image.jpg")
	})
	t.Run("should return error if image object is not valid", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]interface{}{
				"someWrongKey": "some/local/path/image.jpg",
			},
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageObjectInvalid)
		require.Zero(t, found)
	})
	t.Run("should not return anything if only imageUrl is found", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageURLKey: "http://some.url",
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Zero(t, found)
	})
	t.Run("should return error if file has unrecognized extension", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]interface{}{
				localPathKey: "some/local/path/image.txt",
			},
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errFileMustBeImage)
		require.Zero(t, found)
	})
}

const mockImagePath = "./testdata/imageTest.png"

func TestApplyUploadImage(t *testing.T) {
	t.Run("should upload image successfully", func(t *testing.T) {
		mockResp := &marketplace.UploadImageResponse{
			Location: "https://example.org/image.png",
		}
		mockServer := uploadImageMockServer(t, http.StatusOK, mockResp)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := uploadImage(context.Background(), client, mockTenantID, mockImagePath)
		require.NoError(t, err)
		require.Equal(t, "https://example.org/image.png", found)
	})

	t.Run("should return error if image file does not exist", func(t *testing.T) {
		mockServer := uploadImageMockServer(t, http.StatusNoContent, nil)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := uploadImage(context.Background(), client, mockTenantID, "some/path/image.png")
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Zero(t, found)
	})

	t.Run("should return error if companyID is not defined", func(t *testing.T) {
		mockServer := uploadImageMockServer(t, http.StatusNoContent, nil)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := uploadImage(context.Background(), client, "", mockImagePath)
		require.ErrorIs(t, err, errCompanyIDNotDefined)
		require.Zero(t, found)
	})
}

func uploadImageMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	mockImageURI := fmt.Sprintf(uploadImageEndpoint, mockTenantID)
	imageFile, err := os.Open(mockImagePath)
	require.NoError(t, err)
	imageBytes, err := io.ReadAll(imageFile)
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, err)
		require.Equal(t, mockImageURI, r.RequestURI)
		require.Equal(t, http.MethodPost, r.Method)

		foundReqFile, _, err := r.FormFile(multipartFieldName)
		require.NoError(t, err)
		foundReqFileBytes, err := io.ReadAll(foundReqFile)
		require.NoError(t, err)
		require.Equal(t, imageBytes, foundReqFileBytes)

		w.WriteHeader(statusCode)
		resBytes, err := json.Marshal(mockResponse)
		require.NoError(t, err)
		w.Write(resBytes)
	}))
}
