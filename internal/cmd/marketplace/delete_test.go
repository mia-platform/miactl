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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockDeleteObjectID  = "object-id"
	mockDeleteCompanyID = "company-id"
)

func TestDeleteResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DeleteCmd(opts)
		require.NotNil(t, cmd)
	})
}

func deleteByIDMockServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != fmt.Sprintf(deleteMarketplaceEndpointTemplate, mockDeleteCompanyID, mockDeleteObjectID) && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(statusCode)
	}))
}

func TestDeleteItemByObjectId(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		err          bool
	}{
		"valid delete response": {
			server: deleteByIDMockServer(t, http.StatusNoContent),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: false,
		},
		"resource not found": {
			server: deleteByIDMockServer(t, http.StatusNotFound),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
		"internal server error": {
			server: deleteByIDMockServer(t, http.StatusInternalServerError),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
		"unexpected server response error": {
			server: deleteByIDMockServer(t, http.StatusBadGateway),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
		"unexpected server response 2xx": {
			server: deleteByIDMockServer(t, http.StatusAccepted),
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			err = deleteItemByObjectID(client, mockDeleteCompanyID, mockDeleteObjectID)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
