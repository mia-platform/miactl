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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestApplyGetAndValidateImageLocalPath(t *testing.T) {
	t.Run("should return local path if element contains image - YAML", func(t *testing.T) {
		mockItemYAML := []byte(`---
image:
  localPath: ./someImage.png
`)
		mockItem := &marketplace.Item{}
		err := yaml.Unmarshal(mockItemYAML, mockItem)
		require.NoError(t, err)

		found, err := getAndValidateImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Equal(t, found, "./someImage.png")
	})
	t.Run("should throw error with an item that contains both image and imageURL", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]interface{}{
				localPathKey: "some/local/path/image.jpg",
			},
			imageURLKey: "http://some.url",
		}

		found, err := getAndValidateImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageObjKeysConflict)
		require.Zero(t, found)
	})

	t.Run("should return local path if element contains image", func(t *testing.T) {
		mockItemJSON := []byte(`{
			"image": {
				"localPath": "some/local/path/image.jpg"
			}
		}`)
		mockItem := &marketplace.Item{}
		err := yaml.Unmarshal(mockItemJSON, mockItem)
		require.NoError(t, err)

		found, err := getAndValidateImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Equal(t, found, "some/local/path/image.jpg")
	})

	t.Run("should return error if image object is not valid", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]interface{}{
				"someWrongKey": "some/local/path/image.jpg",
			},
		}

		found, err := getAndValidateImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageObjectInvalid)
		require.Zero(t, found)
	})
	t.Run("should not return anything if only imageUrl is found", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageURLKey: "http://some.url",
		}

		found, err := getAndValidateImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Zero(t, found)
	})
}

const mockImagePath = "./testdata/imageTest.png"

type ErrReader struct {
	err error
}

func (er *ErrReader) Read([]byte) (int, error) {
	return 0, er.err
}

func TestApplyReadContentType(t *testing.T) {
	t.Run("should read correct content type", func(t *testing.T) {
		imageFile, err := os.Open(mockImagePath)
		require.NoError(t, err)
		defer imageFile.Close()
		found, err := readContentType(imageFile)
		require.NoError(t, err)
		require.Equal(t, "image/png", found)
	})

	t.Run("should return error if read fails", func(t *testing.T) {
		mockErr := errors.New("testing error")
		found, err := readContentType(
			&ErrReader{
				err: mockErr,
			},
		)
		require.ErrorIs(t, err, mockErr)
		require.Zero(t, found)
	})
}

func TestApplyUploadImage(t *testing.T) {
	t.Run("should upload image successfully", func(t *testing.T) {
		imageFile, err := os.Open(mockImagePath)
		require.NoError(t, err)
		defer imageFile.Close()

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

		found, err := uploadSingleFileWithMultipart(
			context.Background(),
			client,
			mockTenantID,
			"image/png",
			imageFile.Name(),
			imageFile,
			"someItemId",
			"someAssetType",
		)

		require.NoError(t, err)
		require.Equal(t, "https://example.org/image.png", found)
	})

	t.Run("should return error if companyID is not defined", func(t *testing.T) {
		imageFile, err := os.Open(mockImagePath)
		require.NoError(t, err)
		defer imageFile.Close()

		mockServer := uploadImageMockServer(t, http.StatusNoContent, nil)
		defer mockServer.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = mockServer.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := uploadSingleFileWithMultipart(
			context.Background(),
			client,
			"",
			"image/png",
			imageFile.Name(),
			imageFile,
			"someItemId",
			"someAssetType",
		)
		require.ErrorIs(t, err, errCompanyIDNotDefined)
		require.Zero(t, found)
	})
}

func TestValidateImageFile(t *testing.T) {
	t.Run("should return error if content type is not allowed", func(t *testing.T) {
		contentType := "application/javascript"

		err := validateImageContentType(contentType)
		require.ErrorIs(t, err, errFileMustBeImage)
	})

	t.Run("should not return error if content type is png", func(t *testing.T) {
		contentType := "image/png"

		err := validateImageContentType(contentType)
		require.NoError(t, err)
	})
	t.Run("should not return error if content type is jpg", func(t *testing.T) {
		contentType := "image/jpg"

		err := validateImageContentType(contentType)
		require.NoError(t, err)
	})
	t.Run("should not return error if content type is jpeg", func(t *testing.T) {
		contentType := "image/jpeg"

		err := validateImageContentType(contentType)
		require.NoError(t, err)
	})
}

func uploadImageHandler(t *testing.T, w http.ResponseWriter, r *http.Request, statusCode int, mockResponse interface{}) {
	t.Helper()

	mockImageURI := fmt.Sprintf(uploadImageEndpointTemplate, mockTenantID)
	imageFile, err := os.Open(mockImagePath)
	require.NoError(t, err)
	imageBytes, err := io.ReadAll(imageFile)
	require.NoError(t, err)
	require.Equal(t, mockImageURI, r.RequestURI)
	require.Equal(t, http.MethodPost, r.Method)
	require.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

	foundReqFile, _, err := r.FormFile(multipartFieldName)
	require.NoError(t, err)
	foundReqFileBytes, err := io.ReadAll(foundReqFile)
	require.NoError(t, err)
	require.Equal(t, imageBytes, foundReqFileBytes)

	w.WriteHeader(statusCode)
	resBytes, err := json.Marshal(mockResponse)
	require.NoError(t, err)
	w.Write(resBytes)
}

func uploadImageMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadImageHandler(t, w, r, statusCode, mockResponse)
	}))
}
