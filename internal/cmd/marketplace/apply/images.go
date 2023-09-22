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
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

const (
	// uploadImageEndpoint has to be `Sprintf`ed with the companyID
	uploadImageEndpoint = "/api/backend/marketplace/tenants/%s/files"
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
		imageInfoObj, ok := imageInfo.(map[string]interface{})
		if !ok {
			return "", errImageObjectInvalid
		}
		localPath, ok := imageInfoObj[localPathKey]
		if !ok {
			return "", errImageObjectInvalid
		}
		localPathStr, ok := localPath.(string)
		if !ok {
			return "", errImageObjectInvalid
		}
		switch filepath.Ext(localPathStr) {
		case pngExtension, jpegExtension, jpgExtension:
			return localPathStr, nil
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
	if fw, err = createFormFileWithContentType(multipartWriter, multipartFieldName, filepath.Base(imagePath), "image/png"); err != nil {
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

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func createFormFileWithContentType(w *multipart.Writer, fieldname, filename, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			fieldname, filename))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}
