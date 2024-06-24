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

	testCases := map[string]struct {
		companyID string
		server    *httptest.Server
		err       bool
	}{
		"valid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf(listAPIFmt, "company-1"),
				verb: http.MethodGet,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   validBodyString,
			}),
		},
		"invalid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf(listAPIFmt, "company-1"),
				verb: http.MethodGet,
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

func TestE11yClientApply(t *testing.T) {
	testCases := map[string]struct {
		companyID           string
		extensionID         string
		extension           *extensibility.Extension
		expectedExtensionID string
		server              *httptest.Server
		err                 bool
	}{
		"valid response for update": {
			companyID: "company-1",
			extension: &extensibility.Extension{
				ExtensionID: "ext-1",
			},
			expectedExtensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions",
				verb: http.MethodPut,
			}, MockResponse{
				statusCode: http.StatusNoContent,
			}),
		},
		"valid response for insert": {
			companyID:           "company-1",
			extension:           &extensibility.Extension{},
			expectedExtensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions",
				verb: http.MethodPut,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   `{"extensionId":"ext-1"}`,
			}),
		},
		"invalid response": {
			companyID:           "company-1",
			extension:           &extensibility.Extension{},
			expectedExtensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions",
				verb: http.MethodPut,
			}, MockResponse{
				statusCode: http.StatusInternalServerError,
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

			id, err := New(client).Apply(context.TODO(), testCase.companyID, testCase.extension)
			if testCase.err {
				require.Error(t, err)
				require.Zero(t, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedExtensionID, id)
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
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf(deleteAPIFmt, "company-1", "ext-1"),
				verb: http.MethodDelete,
			}, MockResponse{
				statusCode: http.StatusNoContent,
			}),
		},
		"invalid response": {
			companyID:   "company-1",
			extensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf(deleteAPIFmt, "company-1", "ext-1"),
				verb: http.MethodDelete,
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
			activationScopeRequest: ActivationScope{ContextID: "company-1", ContextType: CompanyContext},
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
			activationScopeRequest: ActivationScope{ContextID: "project-1", ContextType: ProjectContext},
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

func TestE11yClientDeactivate(t *testing.T) {
	testCases := map[string]struct {
		companyID                string
		extensionID              string
		deactivationScopeRequest ActivationScope
		server                   *httptest.Server
		err                      bool
	}{
		"valid response": {
			companyID:                "company-1",
			extensionID:              "ext-1",
			deactivationScopeRequest: ActivationScope{ContextID: "company-1", ContextType: CompanyContext},
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-1/company/company-1/activation",
				verb: http.MethodDelete,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
		},
		"valid response for project deactivation": {
			companyID:                "company-1",
			extensionID:              "ext-1",
			deactivationScopeRequest: ActivationScope{ContextID: "project-1", ContextType: ProjectContext},
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-12/project/p-1/activation",
				verb: http.MethodDelete,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
		},
		"invalid response": {
			companyID:   "company-1",
			extensionID: "ext-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/extensibility/tenants/company-1/extensions/ext-1/project/p-1/activation",
				verb: http.MethodDelete,
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

			err = New(client).Deactivate(context.TODO(), testCase.companyID, testCase.extensionID, testCase.deactivationScopeRequest)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
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
