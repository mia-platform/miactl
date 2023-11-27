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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
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
		require.Equal(t,
			fmt.Sprintf(deleteMarketplaceEndpointTemplate, mockDeleteCompanyID, mockDeleteObjectID),
			r.RequestURI,
		)
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(statusCode)
	}))
}

func deleteByItemIDAndVersionMockServer(t *testing.T,
	statusCode int,
	mockItemID, mockVersion string,
	callsCount *int,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t,
			fmt.Sprintf(deleteItemByTupleEndpointTemplate, mockDeleteCompanyID, mockItemID, mockVersion),
			r.RequestURI,
		)
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(statusCode)
		*callsCount++
	}))
}

func TestDeleteItemByObjectId(t *testing.T) {
	mockClientConfig := &client.Config{
		Transport: http.DefaultTransport,
	}
	testCases := map[string]struct {
		server      *httptest.Server
		expectedErr error
	}{
		"valid delete response": {
			server:      deleteByIDMockServer(t, http.StatusNoContent),
			expectedErr: nil,
		},
		"resource not found": {
			server:      deleteByIDMockServer(t, http.StatusNotFound),
			expectedErr: marketplace.ErrItemNotFound,
		},
		"internal server error": {
			server:      deleteByIDMockServer(t, http.StatusInternalServerError),
			expectedErr: errServerDeleteItem,
		},
		"unexpected server response error": {
			server:      deleteByIDMockServer(t, http.StatusBadGateway),
			expectedErr: errServerDeleteItem,
		},
		"unexpected server response 2xx": {
			server:      deleteByIDMockServer(t, http.StatusAccepted),
			expectedErr: errUnexpectedDeleteItem,
		},
		"unexpected server response 4xx": {
			server:      deleteByIDMockServer(t, http.StatusBadRequest),
			expectedErr: errUnexpectedDeleteItem,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			mockClientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(mockClientConfig)
			require.NoError(t, err)
			err = deleteItemByObjectID(
				context.Background(),
				client,
				mockDeleteCompanyID,
				mockDeleteObjectID,
			)
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteItemByItemIDAndVersion(t *testing.T) {
	mockClientConfig := &client.Config{
		Transport: http.DefaultTransport,
	}
	testCases := []struct {
		testName string

		statusCode int

		itemID  string
		version string

		expectedErr   error
		expectedCalls int
	}{
		{
			testName:   "should not return error if deletion is successful",
			statusCode: http.StatusNoContent,
			itemID:     "some-id",
			version:    "1.0.0",

			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			testName: "should return not found error in case the item is not found",
			itemID:   "some-id",
			version:  "1.0.0",

			statusCode: http.StatusNotFound,

			expectedErr:   marketplace.ErrItemNotFound,
			expectedCalls: 1,
		},
		{
			testName: "should return generic error in case the server responds 500",
			itemID:   "some-id",
			version:  "1.0.0",

			statusCode: http.StatusInternalServerError,

			expectedErr:   errServerDeleteItem,
			expectedCalls: 1,
		},
		{
			testName: "should return unexpected response error in case of bad request response",
			itemID:   "some-id",
			version:  "1.0.0",

			statusCode: http.StatusBadRequest,

			expectedErr:   errUnexpectedDeleteItem,
			expectedCalls: 1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			callsCount := new(int)
			*callsCount = 0
			testServer := deleteByItemIDAndVersionMockServer(
				t,
				tt.statusCode,
				tt.itemID,
				tt.version,
				callsCount,
			)
			defer testServer.Close()

			mockClientConfig.Host = testServer.URL
			client, err := client.APIClientForConfig(mockClientConfig)
			require.NoError(t, err)

			err = deleteItemByItemIDAndVersion(
				context.Background(),
				client,
				mockDeleteCompanyID,
				tt.itemID,
				tt.version,
			)

			require.Equal(t, tt.expectedCalls, *callsCount, "did not match number of calls")

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
