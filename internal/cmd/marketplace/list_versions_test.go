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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	listVersionsMockResponseBody = `[
    {
		"name": "Some Awesome Service",
		"description": "The Awesome Service allows to do some amazing stuff.",
		"version": "1.0.0",
		"reference": "655342ce0f991db238fd73e4",
		"security": false,
		"releaseNote": "-",
		"visibility": {
		  "public": true
		}
	},
	{
		"name": "Some Awesome Service v2",
		"description": "The Awesome Service allows to do some amazing stuff.",
		"version": "2.0.0",
		"reference": "655342ce0f991db238fd73e4",
		"security": false,
		"releaseNote": "-",
		"visibility": {
		  "public": true
		}
	}
]`
)

func TestNewListVersionsCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListVersionCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestGetItemVersions(t *testing.T) {
	testCases := []struct {
		testName string

		companyID string
		itemID    string

		statusCode    int
		errorResponse map[string]string

		expected    []marketplace.Release
		expectedErr error
	}{
		{
			testName:   "should return correct result when the endpoint answers 200 OK",
			companyID:  "some-company",
			itemID:     "some-item",
			statusCode: http.StatusOK,
			expected: []marketplace.Release{
				{
					Name:        "Some Awesome Service",
					Description: "The Awesome Service allows to do some amazing stuff.",
					Version:     "1.0.0",
				},
				{
					Name:        "Some Awesome Service v2",
					Description: "The Awesome Service allows to do some amazing stuff.",
					Version:     "2.0.0",
				},
			},
			expectedErr: nil,
		},
		{
			testName:   "should return not found error if item is not found",
			companyID:  "some-company",
			itemID:     "some-item",
			statusCode: http.StatusNotFound,
			errorResponse: map[string]string{
				"error": "Not Found",
			},
			expected:    nil,
			expectedErr: marketplace.ErrItemNotFound,
		},
		{
			testName:   "should return generic error if item is not found",
			companyID:  "some-company",
			itemID:     "some-item",
			statusCode: http.StatusInternalServerError,
			errorResponse: map[string]string{
				"error": "Internal Server Error",
			},
			expected:    nil,
			expectedErr: ErrGenericServerError,
		},
		{
			testName:    "should return error on missing companyID",
			companyID:   "",
			itemID:      "some-item",
			expectedErr: marketplace.ErrMissingCompanyID,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			var expectedBytes []byte
			if testCase.expected != nil {
				expectedBytes = []byte(listVersionsMockResponseBody)
			} else {
				var err error
				expectedBytes, err = json.Marshal(testCase.errorResponse)
				require.NoError(t, err)
			}

			mockServer := buildMockListVersionServer(
				t,
				testCase.statusCode,
				expectedBytes,
				testCase.companyID,
				testCase.itemID,
			)
			defer mockServer.Close()
			client, err := client.APIClientForConfig(&client.Config{
				Transport: http.DefaultTransport,
				Host:      mockServer.URL,
			})
			require.NoError(t, err)

			found, err := getItemVersions(context.TODO(), client, testCase.companyID, testCase.itemID)
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, testCase.expectedErr)
				require.Nil(t, found)
			} else {
				require.NoError(t, err)
				require.Equal(t, &testCase.expected, found)
			}
		})
	}
}

func TestBuildMarketplaceItemVersionList(t *testing.T) {
	testCases := map[string]struct {
		releases         []marketplace.Release
		expectedContains []string
	}{
		"should show all fields": {
			releases: []marketplace.Release{
				{
					Version:     "1.0.0",
					Name:        "Some Awesome Service",
					Description: "The Awesome Service allows to do some amazing stuff.",
				},
			},
			expectedContains: []string{
				"VERSION", "NAME", "DESCRIPTION",
				"1.0.0", "Some Awesome Service", "The Awesome Service allows to do some amazing stuff.",
			},
		},
		"should show - on empty description": {
			releases: []marketplace.Release{
				{
					Version: "1.0.0",
					Name:    "Some Awesome Service",
				},
			},
			expectedContains: []string{
				"VERSION", "NAME", "DESCRIPTION",
				"1.0.0", "Some Awesome Service", "-",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			strBuilder := &strings.Builder{}
			printItemVersionList(&testCase.releases, printer.NewTablePrinter(printer.TablePrinterOptions{WrapLinesDisabled: true}, strBuilder))
			found := strBuilder.String()
			assert.NotZero(t, found)
			for _, expected := range testCase.expectedContains {
				assert.Contains(t, found, expected)
			}
		})
	}
}

func buildMockListVersionServer(t *testing.T, statusCode int, responseBody []byte, companyID, itemID string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(
			t,
			r.RequestURI,
			fmt.Sprintf(listItemVersionsEndpointTemplate, companyID, itemID),
		)
		w.WriteHeader(statusCode)
		w.Write(responseBody)
	}))
}
