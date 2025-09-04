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
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})

	t.Run("should not run command when Console version is lower than 14.1.0", func(t *testing.T) {
		server := httptest.NewServer(unexecutedCmdMockServer(t))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "my-company"
		opts.Endpoint = server.URL

		cmd := ListCmd(opts)
		cmd.SetArgs([]string{"list"})

		err := cmd.Execute()
		require.ErrorIs(t, err, itd.ErrUnsupportedCompanyVersion)
	})
}

func TestBuildMarketplaceItemsList(t *testing.T) {
	type testCase struct {
		name             string
		options          GetItdsOptions
		responseHandler  http.HandlerFunc
		clientConfig     *client.Config
		expectError      bool
		expectedContains []string
	}

	testCases := []testCase{
		{
			name: "private item type definitions",
			options: GetItdsOptions{
				CompanyID: "my-company",
				Public:    false,
			},
			responseHandler: privateItdsHandler(t),
			clientConfig:    &client.Config{Transport: http.DefaultTransport},
			expectError:     false,
			expectedContains: []string{
				"NAME", "DISPLAY NAME", "VISIBILITY", "PUBLISHER", "VERSIONING SUPPORTED",
				"plugin", "Plugin", "tenant", "Test Publisher", "false",
			},
		},
		{
			name: "public and private item type definitions",
			options: GetItdsOptions{
				CompanyID: "my-company",
				Public:    true,
			},
			responseHandler: publicAndPrivateItdsHandler(t),
			clientConfig:    &client.Config{Transport: http.DefaultTransport},
			expectError:     false,
			expectedContains: []string{
				"NAME", "DISPLAY NAME", "VISIBILITY", "PUBLISHER", "VERSIONING SUPPORTED",
				"plugin", "Plugin", "tenant", "Test Publisher", "false",
				"custom-resource", "Infrastructure resource", "console", "Mia-Platform", "true",
				"itd-no-publisher", "ITD no publisher", "console", "-", "false",
			},
		},
	}

	runTestCase := func(t *testing.T, tc testCase) {
		t.Helper()
		server := httptest.NewServer(tc.responseHandler)
		defer server.Close()

		tc.clientConfig.Host = server.URL
		client, err := client.APIClientForConfig(tc.clientConfig)
		require.NoError(t, err)

		strBuilder := &strings.Builder{}
		mockPrinter := printer.NewTablePrinter(printer.TablePrinterOptions{}, strBuilder)
		err = PrintItds(t.Context(), client, tc.options, mockPrinter, listItdEndpoint)
		found := strBuilder.String()
		if tc.expectError {
			assert.Error(t, err)
			assert.Zero(t, found)
		} else {
			assert.NoError(t, err)
			assert.NotZero(t, found)
			for _, expected := range tc.expectedContains {
				assert.Contains(t, found, expected)
			}
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func unexecutedCmdMockServer(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, "/api/version") && r.Method == http.MethodGet {
			_, err := w.Write([]byte(`{"major": "14", "minor":"0"}`))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, fmt.Sprintf("unexpected request: %s", r.URL.Path))
		}
	}
}

func privateItdsHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, listItdEndpoint) &&
			r.Method == http.MethodGet &&
			r.URL.Query().Get("visibility") == "my-company" {
			_, err := w.Write([]byte(privateItdsBodyContent(t)))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, fmt.Sprintf("unexpected request: %s", r.URL.Path))
		}
	}
}

func publicAndPrivateItdsHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, listItdEndpoint) &&
			r.Method == http.MethodGet &&
			r.URL.Query().Get("visibility") == "console,my-company" {
			_, err := w.Write([]byte(publicItdsBodyContent(t)))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, fmt.Sprintf("unexpected request: %s", r.URL.Path))
		}
	}
}

func privateItdsBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"apiVersion": "software-catalog.mia-platform.eu/v1",
			"kind": "item-type-definition",
			"metadata": {
					"namespace": {
							"scope": "tenant",
							"id": "99350849-653a-418c-8a66-545b4b34b619"
					},
					"name": "plugin",
					"visibility": {
							"scope": "tenant",
							"ids": ["99350849-653a-418c-8a66-545b4b34b619"]
					},
					"displayName": "Plugin",
					"publisher": {
							"name": "Test Publisher"
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
			"__v": 9
	  }
	]`
}

func publicItdsBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"apiVersion": "software-catalog.mia-platform.eu/v1",
			"kind": "item-type-definition",
			"metadata": {
					"namespace": {
							"scope": "tenant",
							"id": "99350849-653a-418c-8a66-545b4b34b619"
					},
					"name": "plugin",
					"visibility": {
							"scope": "tenant",
							"ids": ["99350849-653a-418c-8a66-545b4b34b619"]
					},
					"displayName": "Plugin",
					"publisher": {
							"name": "Test Publisher"
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
			"__v": 9
	  },
		{
			"apiVersion": "software-catalog.mia-platform.eu/v1",
			"kind": "item-type-definition",
			"metadata": {
					"namespace": {
							"scope": "tenant",
							"id": "mia-platform"
					},
					"name": "custom-resource",
					"visibility": {
							"scope": "console"
					},
					"displayName": "Infrastructure resource",
					"description": "Defines custom objects beyond the standard Console-supported resources.",
					"documentation": {
							"type": "external",
							"url": "https://docs.mia-platform.eu/docs/software-catalog/items-manifest/infrastructure-resource"
					},
					"maintainers": [
							{
									"name": "Mia-Platform",
									"email": "support@mia-platform.eu"
							}
					],
					"publisher": {
							"name": "Mia-Platform",
							"url": "https://mia-platform.eu/"
					}
			},
			"spec": {
					"type": "custom-resource",
					"scope": "tenant",
					"validation": {
							"mechanism": "json-schema",
							"schema": {}
					},
					"isVersioningSupported": true
			},
			"__v": 0
		},
		{
			"apiVersion": "software-catalog.mia-platform.eu/v1",
			"kind": "item-type-definition",
			"metadata": {
					"namespace": {
							"scope": "tenant",
							"id": "99350849-653a-418c-8a66-545b4b34b619"
					},
					"name": "itd-no-publisher",
					"visibility": {
							"scope": "console"
					},
					"displayName": "ITD no publisher"
			},
			"spec": {
					"type": "plugin",
					"scope": "tenant",
					"validation": {
							"mechanism": "json-schema",
							"schema": {}
					}
			},
			"__v": 1
	  }
	]`
}
