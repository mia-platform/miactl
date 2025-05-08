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
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const MockImagePath = "./testdata/imageTest.png"
const MockTenantID = "mock-tenant-id"

func UploadImageHandler(t *testing.T, w http.ResponseWriter, r *http.Request, statusCode int, mockResponse interface{}) {
	t.Helper()

	mockImageURI := fmt.Sprintf(UploadImageEndpointTemplate, MockTenantID)
	imageFile, err := os.Open(MockImagePath)
	require.NoError(t, err)
	imageBytes, err := io.ReadAll(imageFile)
	require.NoError(t, err)
	require.Equal(t, mockImageURI, r.RequestURI)
	require.Equal(t, http.MethodPost, r.Method)
	require.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

	foundReqFile, _, err := r.FormFile(MultipartFieldName)
	require.NoError(t, err)
	foundReqFileBytes, err := io.ReadAll(foundReqFile)
	require.NoError(t, err)
	require.Equal(t, imageBytes, foundReqFileBytes)

	w.WriteHeader(statusCode)
	resBytes, err := json.Marshal(mockResponse)
	require.NoError(t, err)
	//nolint:errcheck // ignore error check because function is for testing purposes
	w.Write(resBytes)
}
