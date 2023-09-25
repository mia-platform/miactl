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
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

const (
	// uploadImageEndpoint has to be `Sprintf`ed with the companyID
	uploadImageEndpoint = "/api/backend/marketplace/tenants/%s/files"
	multipartFieldName  = "marketplace_image"

	localPathKey = "localPath"

	jpegMimeType = "image/jpeg"
	jpgMimeType  = "image/jpg"
	pngMimeType  = "image/png"
)

var (
	errImageURLConflict   = errors.New(`both "image" and "imageUrl" found in the item, only one is admitted`)
	errImageObjectInvalid = errors.New("the image object is not valid")

	errFileMustBeImage = errors.New("the file must a jpeg, png or gif image")
)

// getAndValidateImageLocalPath looks for an imageKey in the Marketplace item, if found it returns the local path
func getAndValidateImageLocalPath(item *marketplace.Item, imageKey, imageURLKey string) (string, error) {
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
		return localPathStr, nil
	}

	return "", nil
}

func readContentType(fileContents io.Reader) (string, error) {
	// DetectContentType only needs the first 512 bytes
	headerBytes := make([]byte, 512)
	_, err := fileContents.Read(headerBytes)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(headerBytes)
	return contentType, nil
}

func validateImageContentType(contentType string) error {
	switch contentType {
	case pngMimeType, jpgMimeType, jpegMimeType:
		return nil
	}
	return errFileMustBeImage
}

func buildUploadImageReq(imageMimeType, fileName string, fileContents io.Reader) (string, []byte, error) {
	var bodyBuffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&bodyBuffer)
	formFileWriter, err := createFormFileWithContentType(multipartWriter, multipartFieldName, fileName, imageMimeType)
	if err != nil {
		return "", nil, err
	}
	if _, err = io.Copy(formFileWriter, fileContents); err != nil {
		return "", nil, err
	}
	multipartWriter.Close()

	reqContentType := multipartWriter.FormDataContentType()
	bodyBytes := bodyBuffer.Bytes()

	return reqContentType, bodyBytes, nil
}

func uploadImageFileAndGetURL(ctx context.Context, client *client.APIClient, restConfig *client.Config, filePath string) (string, error) {
	imageFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	contentType, err := readContentType(imageFile)
	if err != nil {
		return "", err
	}
	if err = validateImageContentType(contentType); err != nil {
		return "", err
	}

	// we need to go back to start, as the file needs to be re-read later
	if _, err := imageFile.Seek(0, 0); err != nil {
		return "", err
	}

	imageURL, err := uploadSingleFileWithMultipart(ctx, client, restConfig.CompanyID, contentType, imageFile.Name(), imageFile)
	if err != nil {
		return "", err
	}
	return imageURL, nil
}

// uploadSingleFileWithMultipart uploads the given Reader as a single multipart file
// the part will also be given a filename and a contentType
func uploadSingleFileWithMultipart(ctx context.Context, client *client.APIClient, companyID, fileMimeType, fileName string, fileContents io.Reader) (string, error) {
	if companyID == "" {
		return "", errCompanyIDNotDefined
	}

	contentType, bodyBytes, err := buildUploadImageReq(fileMimeType, fileName, fileContents)
	if err != nil {
		return "", nil
	}

	resp, err := client.Post().
		SetHeader("Content-Type", contentType).
		APIPath(fmt.Sprintf(uploadImageEndpoint, companyID)).
		Body(bodyBytes).
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

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// createFormFileWithContentType reimplements multipart.CreateFormFile(): https://cs.opensource.google/go/go/+/refs/tags/go1.21.1:src/mime/multipart/writer.go;l=140
// It adds the possibility to set an arbitrary contentType MIME header to the file.
// The latter would be otherwise defaulted to application/octet-stream, which is is not accepted by the Mia-Platform backend endpoint, because it needs to know the file type.
// For this reason we need to reimplement the function with this modification, replicating also the Content-Disposition build, until this proposal lands into Go's standard library
// https://github.com/golang/go/issues/46771
func createFormFileWithContentType(w *multipart.Writer, fieldname, filename, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}
