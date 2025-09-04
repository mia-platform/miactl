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

package itd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/stretchr/testify/require"
)

func TestPutCommand(t *testing.T) {
	t.Run("test post run - shows deprecated command message", func(t *testing.T) {
		server := httptest.NewServer(unexecutedCmdMockServer(t))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "company-id"
		opts.Endpoint = server.URL

		cmd := PutCmd(opts)
		cmd.SetArgs([]string{"apply", "-f", "testdata/validItem1.json"})

		err := cmd.Execute()
		require.ErrorIs(t, err, itd.ErrUnsupportedCompanyVersion)
	})
}

var mockTenantID = "mock-tenant-id"
var mockURI = "/api/tenants/" + mockTenantID + "/marketplace/item-type-definitions/"
var mockFilePath = "./testdata/validItd.json"

func TestApplyApplyResourceCmd(t *testing.T) {
	// testdata/validItd.json with __v + 1
	mockResponseJSON := `{
		"apiVersion": "software-catalog.mia-platform.eu/v1",
		"kind": "item-type-definition",
		"metadata": {
			"namespace": {
				"scope": "tenant",
				"id": "some-company-id"
			},
			"name": "plugin",
			"visibility": {
				"scope": "tenant",
				"ids": [
					"some-company-id"
				]
			},
			"displayName": "Plugin",
			"tags": [
				"prova"
			],
			"maintainers": [
				{
					"name": "ok"
				}
			],
			"publisher": {
				"name": "publisher-name"
			}
		},
		"spec": {
			"type": "plugin",
			"scope": "tenant",
			"validation": {
				"mechanism": "json-schema",
				"schema": {}
			}
		},
		"__v": 3
	}`

	var mockResponse itd.GenericItemTypeDefinition
	err := json.Unmarshal([]byte(mockResponseJSON), &mockResponse)
	require.NoError(t, err)

	t.Run("should return response when is 200 OK", func(t *testing.T) {
		server := putMockServer(t, http.StatusOK, mockResponse)
		defer server.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := putItemTypeDefinition(
			t.Context(),
			client,
			mockTenantID,
			mockFilePath,
		)

		require.NoError(t, err)
		require.Equal(t, &mockResponse, found)
	})

	t.Run("should return err if response is a http error", func(t *testing.T) {
		mockErrorResponse := map[string]interface{}{
			"message":    "You are not allowed to perform the request!",
			"statusCode": http.StatusForbidden,
			"error":      "Forbidden",
		}
		server := putMockServer(t, http.StatusForbidden, mockErrorResponse)
		defer server.Close()
		clientConfig := &client.Config{
			Transport: http.DefaultTransport,
		}
		clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(clientConfig)
		require.NoError(t, err)

		found, err := putItemTypeDefinition(
			t.Context(),
			client,
			mockTenantID,
			mockFilePath,
		)

		require.EqualError(t, err, "You are not allowed to perform the request!")
		require.Nil(t, found)
	})
}

func putMockServer(t *testing.T, statusCode int, mockResponse interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		putRequestHandler(t, w, r, statusCode, mockResponse)
	}))
}

func putRequestHandler(t *testing.T, w http.ResponseWriter, r *http.Request, statusCode int, mockResponse interface{}) {
	t.Helper()

	require.Equal(t, mockURI, r.RequestURI)
	require.Equal(t, http.MethodPut, r.Method)

	w.WriteHeader(statusCode)
	resBytes, err := json.Marshal(mockResponse)
	require.NoError(t, err)
	w.Write(resBytes)
}
