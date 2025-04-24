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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
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

	t.Run("test post run - shows deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(deleteItemCommandMockServer(t, `{"major": "14", "minor":"1"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = mockDeleteCompanyID
		opts.Endpoint = server.URL

		cmd := DeleteCmd(opts)
		cmd.SetArgs([]string{"delete", "--item-id", "some-item-id", "--version", "1.0.0"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "item deleted successfully")

		outputErr := buffer.String()
		assert.Contains(t, outputErr, "The command you are using is deprecated. Please use 'miactl catalog' instead.")
	})

	t.Run("test post run - does not show deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(deleteItemCommandMockServer(t, `{"major": "13", "minor":"5"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = mockDeleteCompanyID
		opts.Endpoint = server.URL

		cmd := DeleteCmd(opts)
		cmd.SetArgs([]string{"delete", "--item-id", "some-item-id", "--version", "1.0.0"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "item deleted successfully")

		outputErr := buffer.String()
		assert.Equal(t, outputErr, "")
	})
}

func deleteItemCommandMockServer(t *testing.T, consoleVersionResponse string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/backend/marketplace/tenants/company-id/resources/some-item-id/versions/1.0.0":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
			}
		case "/api/version":
			if r.Method == http.MethodGet {
				_, err := w.Write([]byte(consoleVersionResponse))
				require.NoError(t, err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, fmt.Sprintf("unexpected request: %s", r.URL.Path))
		}
	}
}

func deleteByIDMockServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t,
			fmt.Sprintf(deleteItemEndpointTemplate, mockDeleteCompanyID, mockDeleteObjectID),
			r.RequestURI,
		)
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(statusCode)
		if statusCode != http.StatusNoContent {
			w.Write([]byte(`
			{
				"message": "some error message"
			}
			`))
		}
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
		if statusCode != http.StatusNoContent {
			w.Write([]byte(`
			{
				"message": "some error message"
			}
			`))
		}
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
				t.Context(),
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
				t.Context(),
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
