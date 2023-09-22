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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

const (
	// uploadImageEndpoint has to be `Sprintf`ed with the companyID
	uploadImageEndpoint = "/api/marketplace/tenants/%s/files"
	multipartFieldName  = "marketplace_image"

	localPathKey = "localPath"

	jpegExtension = ".jpeg"
	jpgExtension  = ".jpg"
	pngExtension  = ".png"
	gifExtension  = ".gif"
)

var (
	errImageURLConflict   = errors.New(`both "image" and "imageUrl" found in the item, only one is admitted`)
	errImageObjectInvalid = errors.New("the image object is not valid")

	errFileMustBeImage = errors.New("the file must a jpeg, png or gif image")
)

// validateAndGetImageLocalPath looks for an imageKey in the Item, if found it returns the local path
func validateAndGetImageLocalPath(item *marketplace.Item, imageKey, imageURLKey string) (string, error) {
	_, imageURLExists := (*item)[imageURLKey]
	imageInfo, imageExists := (*item)[imageKey]
	if imageExists && imageURLExists {
		return "", errImageURLConflict
	}

	if imageExists {
		localPath, ok := imageInfo.(map[string]string)[localPathKey]
		if !ok {
			return "", errImageObjectInvalid
		}
		switch filepath.Ext(localPath) {
		case pngExtension, jpegExtension, jpgExtension:
			return localPath, nil
		default:
			return "", errFileMustBeImage
		}
	}

	return "", nil
}

// uploadImage uploads an image and returns the URL
func uploadImage(ctx context.Context, client *client.APIClient, companyID, imagePath string) (string, error) {
	if companyID == "" {
		return "", errCompanyIDNotDefined
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var bodyBuffer bytes.Buffer
	var fw io.Writer
	multipartWriter := multipart.NewWriter(&bodyBuffer)
	if fw, err = multipartWriter.CreateFormFile(multipartFieldName, filepath.Base(imagePath)); err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return "", err
	}
	multipartWriter.Close()

	resp, err := client.Post().
		SetHeader("Content-Type", multipartWriter.FormDataContentType()).
		APIPath(fmt.Sprintf(uploadImageEndpoint, companyID)).
		Body(bodyBuffer.Bytes()).
		Do(ctx)
	if err != nil {
		return "", err
	}
	if err := resp.Error(); err != nil {
		return "", err
	}

	uploadResp := &marketplace.UploadImageResponse{}

	err = resp.ParseResponse(uploadResp)
	if err != nil {
		return "", err
	}

	return uploadResp.Location, nil
}
