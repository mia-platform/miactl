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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
)

const (
	mockDeleteCompanyID = "company-id"
)

func TestDeleteResourceCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := DeleteCmd(opts)
		require.NotNil(t, cmd)
	})

	t.Run("should not run command when Console version is lower than 14.1.0", func(t *testing.T) {
		server := httptest.NewServer(unexecutedCmdMockServer(t))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = mockDeleteCompanyID
		opts.Endpoint = server.URL

		cmd := DeleteCmd(opts)
		cmd.SetArgs([]string{"delete", "--name", "some-item-id"})

		err := cmd.Execute()
		require.ErrorIs(t, err, itd.ErrUnsupportedCompanyVersion)
	})
}

func deleteByItemNameMockServer(t *testing.T,
	statusCode int,
	mockName string,
	callsCount *int,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t,
			fmt.Sprintf(deleteItdEndpoint, mockDeleteCompanyID, mockName),
			r.RequestURI,
		)
		assert.Equal(t, http.MethodDelete, r.Method)
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

func TestDeleteItemByItemIDAndVersion(t *testing.T) {
	mockClientConfig := &client.Config{
		Transport: http.DefaultTransport,
	}
	testCases := []struct {
		testName string

		statusCode int

		name string

		expectedErr   error
		expectedCalls int
	}{
		{
			testName:   "should not return error if deletion is successful",
			statusCode: http.StatusNoContent,
			name:       "plugin",

			expectedErr:   nil,
			expectedCalls: 1,
		},
		{
			testName: "should return not found error in case the item is not found",
			name:     "plugin",

			statusCode: http.StatusNotFound,

			expectedErr:   itd.ErrItemNotFound,
			expectedCalls: 1,
		},
		{
			testName: "should return generic error in case the server responds 500",
			name:     "plugin",

			statusCode: http.StatusInternalServerError,

			expectedErr:   ErrServerDeleteItem,
			expectedCalls: 1,
		},
		{
			testName: "should return unexpected response error in case of bad request response",
			name:     "plugin",

			statusCode: http.StatusBadRequest,

			expectedErr:   ErrUnexpectedDeleteItem,
			expectedCalls: 1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			callsCount := new(int)
			*callsCount = 0
			testServer := deleteByItemNameMockServer(
				t,
				tt.statusCode,
				tt.name,
				callsCount,
			)
			defer testServer.Close()

			mockClientConfig.Host = testServer.URL
			client, err := client.APIClientForConfig(mockClientConfig)
			require.NoError(t, err)

			err = deleteITD(
				t.Context(),
				client,
				mockDeleteCompanyID,
				tt.name,
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
