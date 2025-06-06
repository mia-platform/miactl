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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
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

	t.Run("test post run - shows deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(listVersionsCommandHandler(t, `{"major": "14", "minor":"0"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "my-company"
		opts.Endpoint = server.URL

		cmd := ListVersionCmd(opts)
		cmd.SetArgs([]string{"list-versions", "--item-id", "item-id"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "  VERSION  NAME                     DESCRIPTION                                           \n\n  1.0.0    Some Awesome Service     The Awesome Service allows to do some amazing stuff.  \n  2.0.0    Some Awesome Service v2  The Awesome Service allows to do some amazing stuff.  \n")

		outputErr := buffer.String()
		assert.Contains(t, outputErr, "The command you are using is deprecated. Please use 'miactl catalog' instead.")
	})

	t.Run("test post run - does not show deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(listVersionsCommandHandler(t, `{"major": "13", "minor":"5"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "my-company"
		opts.Endpoint = server.URL

		cmd := ListVersionCmd(opts)
		cmd.SetArgs([]string{"list-versions", "--item-id", "item-id"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "  VERSION  NAME                     DESCRIPTION                                           \n\n  1.0.0    Some Awesome Service     The Awesome Service allows to do some amazing stuff.  \n  2.0.0    Some Awesome Service v2  The Awesome Service allows to do some amazing stuff.  \n")

		outputErr := buffer.String()
		assert.Equal(t, outputErr, "")
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
			expectedErr: commonMarketplace.ErrGenericServerError,
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

			found, err := commonMarketplace.GetItemVersions(t.Context(), client, listItemVersionsEndpointTemplate, testCase.companyID, testCase.itemID)
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
			commonMarketplace.PrintItemVersionList(&testCase.releases, printer.NewTablePrinter(printer.TablePrinterOptions{WrapLinesDisabled: true}, strBuilder))
			found := strBuilder.String()
			assert.NotZero(t, found)
			for _, expected := range testCase.expectedContains {
				assert.Contains(t, found, expected)
			}
		})
	}
}

func listVersionsCommandHandler(t *testing.T, consoleVersionResponse string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/backend/marketplace/tenants/my-company/resources/item-id/versions":
			if r.Method == http.MethodGet {
				_, err := w.Write([]byte(listVersionsMockResponseBody))
				require.NoError(t, err)
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
