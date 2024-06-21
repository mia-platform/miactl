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

package extensions

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/stretchr/testify/require"
)

func TestE11yClientList(t *testing.T) {
	testCases := map[string]struct {
		companyID string
		server    *httptest.Server
		err       bool
	}{
		"valid response": {
			companyID: "company-1",
			server:    mockListServer(t, true, "company-1"),
		},
		"invalid response": {
			companyID: "company-1",
			server:    mockListServer(t, false, "company-1"),
			err:       true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			clientConfig := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}

			client, err := client.APIClientForConfig(clientConfig)
			require.NoError(t, err)

			data, err := New(client).List(context.TODO(), testCase.companyID)
			if testCase.err {
				require.Error(t, err)
				require.Nil(t, data)
			} else {
				require.NoError(t, err)
				require.Equal(t, []*extensibility.Extension{
					{
						ExtensionID: "ext-1",
						Name:        "Extension 1",
						Description: "Description 1",
					},
					{
						ExtensionID: "ext-2",
						Name:        "Extension 2",
						Description: "Description 2",
					},
				}, data)
			}
		})
	}
}

func TestE11yClientDelete(t *testing.T) {
	testCases := map[string]struct {
		companyID   string
		extensionID string
		server      *httptest.Server
		err         bool
	}{
		"valid response": {
			companyID:   "company-1",
			extensionID: "ext-1",
			server:      mockDeleteServer(t, true, "company-1", "ext-1"),
		},
		"invalid response": {
			companyID:   "company-1",
			extensionID: "ext-1",
			server:      mockDeleteServer(t, false, "company-1", "ext-1"),
			err:         true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			clientConfig := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}

			client, err := client.APIClientForConfig(clientConfig)
			require.NoError(t, err)

			err = New(client).Delete(context.TODO(), testCase.companyID, testCase.extensionID)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestE11yClientActivate(t *testing.T) {
	testCases := map[string]struct {
		companyID              string
		extensionID            string
		activationScopeRequest ActivationScope
		server                 *httptest.Server
		err                    bool
	}{
		"valid response": {
			companyID:              "company-1",
			extensionID:            "ext-1",
			activationScopeRequest: ActivationScope{ContextID: "company-1", ContextType: "company"},
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-1/activation",
				verb: http.MethodPost,
				body: `{"contextId":"company-1","contextType":"company"}`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
		},
		"valid response for project activation": {
			companyID:              "company-1",
			extensionID:            "ext-1",
			activationScopeRequest: ActivationScope{ContextID: "project-1", ContextType: "project"},
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-1/activation",
				verb: http.MethodPost,
				body: `{"contextId":"project-1","contextType":"project"}`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
		},
		"invalid response": {
			companyID:   "company-1",
			extensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-1/activation",
				verb: http.MethodPost,
			}, MockResponse{
				statusCode: http.StatusInternalServerError,
				err:        true,
			}),
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			clientConfig := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}

			client, err := client.APIClientForConfig(clientConfig)
			require.NoError(t, err)

			err = New(client).Activate(context.TODO(), testCase.companyID, testCase.extensionID, testCase.activationScopeRequest)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mockListServer(t *testing.T, validResponse bool, expectedCompanyID string) *httptest.Server {
	t.Helper()
	validBodyString := `[
	{
		"extensionId": "ext-1",
		"name": "Extension 1",
		"description": "Description 1"
	},
	{
		"extensionId": "ext-2",
		"name": "Extension 2",
		"description": "Description 2"
	}
]`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != fmt.Sprintf(listAPIFmt, expectedCompanyID) && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		w.WriteHeader(http.StatusOK)
		if validResponse {
			w.Write([]byte(validBodyString))
			return
		}
		w.Write([]byte("invalid json"))
	}))
}

func mockDeleteServer(t *testing.T, validResponse bool, expectedCompanyID, expectedExtensionID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != fmt.Sprintf(deleteAPIFmt, expectedCompanyID, expectedExtensionID) && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
			return
		}
		if validResponse {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"some error","message":"some message"}`))
	}))
}

type MockResponse struct {
	statusCode int
	respBody   string
	err        bool
}

type ExpectedRequest struct {
	path string
	verb string
	body string
}

func mockServer(t *testing.T, req ExpectedRequest, resp MockResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != req.path && r.Method != req.verb {
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, fmt.Sprintf("unsupported call: %s - wanted: %s", r.RequestURI, req.path))
			return
		}

		if req.body != "" {
			foundBody, err := io.ReadAll(r.Body)
			if err != nil {
				require.Fail(t, fmt.Sprintf("failed req body read: %s", err.Error()))
			}
			require.Equal(t, req.body, strings.TrimSuffix(string(foundBody), "\n"))
		}

		w.WriteHeader(resp.statusCode)
		if resp.err {
			w.Write([]byte(`{"error":"some error","message":"some message"}`))
			return
		}
		w.Write([]byte(resp.respBody))
	}))
}
