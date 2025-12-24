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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/printer"
)

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := ListCmd(opts)
		require.NotNil(t, cmd)
	})

	t.Run("test post run - shows deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(listCommandHandler(t, `{"major": "14", "minor":"0"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "my-company"
		opts.Endpoint = server.URL

		cmd := ListCmd(opts)
		cmd.SetArgs([]string{"list"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "OBJECT ID                 ITEM ID               NAME                  TYPE    COMPANY ID       \n\n  43774c07d09ac6996ecfb3ef  space-travel-service  Space Travel Service  plugin  my-company       \n  43774c07d09ac6996ecfb3eg  a-public-service      A public service      plugin  another-company  \n")

		outputErr := buffer.String()
		assert.Contains(t, outputErr, "The command you are using is deprecated. Please use 'miactl catalog' instead.")
	})

	t.Run("test post run - does not show deprecated command message", func(t *testing.T) {
		storeStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		server := httptest.NewServer(listCommandHandler(t, `{"major": "13", "minor":"2"}`))
		defer server.Close()

		opts := clioptions.NewCLIOptions()
		opts.CompanyID = "my-company"
		opts.Endpoint = server.URL

		cmd := ListCmd(opts)
		cmd.SetArgs([]string{"list"})

		buffer := bytes.NewBuffer([]byte{})
		cmd.SetErr(buffer)

		err := cmd.Execute()
		require.NoError(t, err)

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = storeStdout
		assert.Contains(t, string(out), "OBJECT ID                 ITEM ID               NAME                  TYPE    COMPANY ID       \n\n  43774c07d09ac6996ecfb3ef  space-travel-service  Space Travel Service  plugin  my-company       \n  43774c07d09ac6996ecfb3eg  a-public-service      A public service      plugin  another-company  \n")

		outputErr := buffer.String()
		assert.Empty(t, outputErr)
	})
}

func TestBuildMarketplaceItemsList(t *testing.T) {
	type testCase struct {
		name             string
		options          commonMarketplace.GetMarketplaceItemsOptions
		responseHandler  http.HandlerFunc
		clientConfig     *client.Config
		expectError      bool
		expectedContains []string
	}

	testCases := []testCase{
		{
			name: "private company marketplace",
			options: commonMarketplace.GetMarketplaceItemsOptions{
				CompanyID: "my-company",
				Public:    false,
			},
			responseHandler: privateCompanyMarketplaceHandler(t),
			clientConfig:    &client.Config{Transport: http.DefaultTransport},
			expectError:     false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE", "COMPANY ID",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin", "my-company",
			},
		},
		{
			name: "wrong payload",
			options: commonMarketplace.GetMarketplaceItemsOptions{
				CompanyID: "my-company",
				Public:    false,
			},
			responseHandler:  wrongPayloadHandler(t),
			clientConfig:     &client.Config{Transport: http.DefaultTransport},
			expectError:      true,
			expectedContains: []string{},
		},
		{
			name: "public marketplace with company set",
			options: commonMarketplace.GetMarketplaceItemsOptions{
				CompanyID: "my-company",
				Public:    true,
			},
			responseHandler: privateAndPublicMarketplaceHandler(t),
			clientConfig:    &client.Config{Transport: http.DefaultTransport},
			expectError:     false,
			expectedContains: []string{
				"ID", "ITEM ID", "NAME", "TYPE", "COMPANY ID",
				"43774c07d09ac6996ecfb3ef", "space-travel-service", "Space Travel Service", "plugin", "my-company",
				"43774c07d09ac6996ecfb3eg", "a-public-service", "A public service", "plugin", "another-company",
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
		err = commonMarketplace.PrintMarketplaceItems(t.Context(), client, tc.options, mockPrinter, listMarketplaceEndpoint)
		found := strBuilder.String()
		if tc.expectError {
			assert.Error(t, err)
			assert.Empty(t, found)
		} else {
			assert.NoError(t, err)
			assert.NotEmpty(t, found)
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

func listCommandHandler(t *testing.T, consoleVersionResponse string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/backend/marketplace/":
			if r.Method == http.MethodGet {
				_, err := w.Write([]byte(marketplaceItemsBodyContent(t)))
				require.NoError(t, err)
			}
		case "/api/version":
			if r.Method == http.MethodGet {
				_, err := w.Write([]byte(consoleVersionResponse))
				require.NoError(t, err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request: "+r.URL.Path)
		}
	}
}

func privateAndPublicMarketplaceHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
			r.Method == http.MethodGet &&
			r.URL.Query().Get("includeTenantId") == "my-company" {
			_, err := w.Write([]byte(marketplaceItemsBodyContent(t)))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request: "+r.URL.Path)
		}
	}
}

func privateCompanyMarketplaceHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
			r.Method == http.MethodGet &&
			r.URL.Query().Get("tenantId") == "my-company" {
			_, err := w.Write([]byte(marketplacePrivateCompanyBodyContent(t)))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request: "+r.URL.Path)
		}
	}
}

func wrongPayloadHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.URL.Path, "/api/backend/marketplace/") &&
			r.Method == http.MethodGet &&
			r.URL.Query().Get("tenantId") == "my-company" {
			_, err := w.Write([]byte("{}")) // Incorrect payload
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, "unexpected request: "+r.URL.Path)
		}
	}
}

func marketplacePrivateCompanyBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"_id": "43774c07d09ac6996ecfb3ef",
			"name": "Space Travel Service",
			"itemId": "space-travel-service",
			"tenantId": "my-company",
			"description": "This service provides a REST API to book your next journey to space!",
			"type": "plugin",
			"imageUrl": "/v2/files/download/space.png",
			"supportedByImageUrl": "/v2/files/download/23b12dd9-43a6-4920-cb2d-2a809d946851.png",
			"supportedBy": "My-Company",
			"category": {
				"id": "auth",
				"label": "Core Plugins - Travel"
			},
			"repositoryUrl": "https://git.com/plugins/core/space-travel-service",
			"documentation": {
				"type": "externalLink",
				"url": "https://docs.my-company.eu/docs/space-travel-service/overview"
			}
		}
	]`
}

func marketplaceItemsBodyContent(t *testing.T) string {
	t.Helper()

	return `[
		{
			"_id": "43774c07d09ac6996ecfb3ef",
			"name": "Space Travel Service",
			"itemId": "space-travel-service",
			"tenantId": "my-company",
			"description": "This service provides a REST API to book your next journey to space!",
			"type": "plugin",
			"imageUrl": "/v2/files/download/space.png",
			"supportedByImageUrl": "/v2/files/download/23b12dd9-43a6-4920-cb2d-2a809d946851.png",
			"supportedBy": "My-Company",
			"category": {
				"id": "auth",
				"label": "Core Plugins - Travel"
			},
			"repositoryUrl": "https://git.com/plugins/core/space-travel-service",
			"documentation": {
				"type": "externalLink",
				"url": "https://docs.my-company.eu/docs/space-travel-service/overview"
			}
		},
		{
			"_id": "43774c07d09ac6996ecfb3eg",
			"name": "A public service",
			"itemId": "a-public-service",
			"tenantId": "another-company",
			"description": "This service provides a REST API to book your next journey to space!",
			"type": "plugin",
			"imageUrl": "/v2/files/download/space.png",
			"supportedByImageUrl": "/v2/files/download/23b12dd9-43a6-4920-cb2d-2a809d946851.png",
			"category": {
				"id": "auth",
				"label": "Core Plugins - Travel"
			},
			"repositoryUrl": "https://git.com/plugins/core/space-travel-service",
			"documentation": {
				"type": "externalLink",
				"url": "https://docs.my-company.eu/docs/space-travel-service/overview"
			},
			"visibility": {
				"public": true
			}
		}
	]`
}
