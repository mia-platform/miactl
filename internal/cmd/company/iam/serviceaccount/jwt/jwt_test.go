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

package jwt

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFromKey(t *testing.T) {
	key, err := generateRSAKey()
	assert.NoError(t, err)

	payload := requestFromKey("testName", resources.IAMRoleCompanyOwner, key)
	assert.Equal(t, "testName", payload.Name)
	assert.Equal(t, resources.IAMRoleCompanyOwner, payload.Role)
	assert.Equal(t, "sig", payload.PublicKey.Use)
	assert.Equal(t, "RSA", payload.PublicKey.Type)
	assert.Equal(t, "RSA256", payload.PublicKey.Algorithm)
	assert.NotEmpty(t, payload.PublicKey.Modulus)
	assert.Equal(t, "AQAB", payload.PublicKey.Exponent)
}

func TestCreateServiceAccount(t *testing.T) {
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		role      resources.IAMRole
		expectErr bool
	}{
		"create successul": {
			server:    testServer(t),
			companyID: "company",
			role:      resources.IAMRoleGuest,
		},
		"wrong role": {
			server:    testServer(t),
			companyID: "unused",
			role:      resources.IAMRole("wrong"),
			expectErr: true,
		},
		"remote error": {
			server:    testServer(t),
			companyID: "error",
			role:      resources.IAMRoleGuest,
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.server
			defer server.Close()
			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)
			response, err := createJWTServiceAccount(context.TODO(), client, "foo", testCase.companyID, testCase.role)
			if testCase.expectErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func TestSaveCredentials(t *testing.T) {
	testBuffer := bytes.NewBuffer([]byte{})
	testCredentials := &resources.JWTServiceAccountJSON{
		Type:           "type",
		KeyID:          "key-id",
		PrivateKeyData: "data",
		ClientID:       "client-id",
	}
	expectedString := `Service account created, save the following json for later uses:
{
	"type": "type",
	"key-id": "key-id",
	"private-key-data": "data",
	"client-id": "client-id"
}
`
	err := saveCredentialsIfNeeded(testCredentials, "", testBuffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedString, testBuffer.String())
	testBuffer.Reset()

	testFile := filepath.Join(t.TempDir(), "file.json")
	expectedString = fmt.Sprintf("Service account created, credentials saved in %s\n", testFile)
	err = saveCredentialsIfNeeded(testCredentials, testFile, testBuffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedString, testBuffer.String())
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(companyServiceAccountsEndpointTemplate, "company"):
			body := &resources.ServiceAccount{
				ClientID:         "client-id",
				ClientIDIssuedAt: 0,
			}
			data, err := resources.EncodeResourceToJSON(body)
			require.NoError(t, err)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(companyServiceAccountsEndpointTemplate, "error"):
			body := &resources.APIError{
				Message:    "error",
				StatusCode: 400,
			}
			data, err := resources.EncodeResourceToJSON(body)
			require.NoError(t, err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))

	return server
}
