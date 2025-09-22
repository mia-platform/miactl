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

package rules

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientListTenantRules(t *testing.T) {
	validBodyString := `[{
	"tenantId": "company-1",
	"configurationManagement": {
		"saveChangesRules": [
			{
				"roleIds": ["maintainer"],
				"disallowedRuleSet": [
					{"jsonPath": "$.services.*.description"},
					{"jsonPath": "$.services", "processingOptions": {"actions": ["create", "delete"]}}
				]
			},
			{
				"roleIds": ["developer"],
				"disallowedRuleSet": [
					{"ruleId": "endpoint.security.edit"}
				]
			},
			{
				"roleIds": ["some-role"],
				"allowedRuleSet": [
					{ "jsonPath": "$.endpoints.*.public" },
					{ "jsonPath": "$.secrets", "processingOptions": { "actions": ["create", "delete"], "primaryKey": "clientType" }}
				]
			}
		]
	}
}]`

	validBodyStringBeforeV14_1 := `[{
	"tenantId": "company-1",
	"configurationManagement": {
		"saveChangesRules": [
			{
				"roleIds": ["maintainer"],
				"disallowedRuleSet": [
					{"jsonPath": "$.services.*.description"},
					{"jsonPath": "$.services", "processingOptions": {"action": "create"}}
				]
			},
			{
				"roleIds": ["developer"],
				"disallowedRuleSet": [
					{"ruleId": "endpoint.security.edit"}
				]
			}
		]
	}
}]`

	expecteRes := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
		{
			RoleIDs: []string{"some-role"},
			AllowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.endpoints.*.public"},
				{JSONPath: "$.secrets", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}, PrimaryKey: "clientType"}},
			},
		},
	}

	expecteResBeforeV14_1 := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Action: "create"}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
	}

	testCases := map[string]struct {
		companyID      string
		server         *httptest.Server
		err            bool
		expectedResult []*rulesentities.SaveChangesRules
	}{
		"valid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/tenants/?search=" + "company-1",
				verb: http.MethodGet,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   validBodyString,
			}),
			expectedResult: expecteRes,
		},
		"valid response - before Console V14.1": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/tenants/?search=" + "company-1",
				verb: http.MethodGet,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   validBodyStringBeforeV14_1,
			}),
			expectedResult: expecteResBeforeV14_1,
		},
		"invalid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/tenants/?search=" + "company-1",
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

			data, err := New(client).ListTenantRules(t.Context(), testCase.companyID)
			if testCase.err {
				require.Error(t, err)
				require.Nil(t, data)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, data)
			}
		})
	}
}

func TestClientListProjectRules(t *testing.T) {
	validBodyString := `{
	"_id": "myproject",
	"tenantId": "company-1",
	"configurationManagement": {
		"saveChangesRules": [
			{
				"roleIds": ["maintainer"],
				"disallowedRuleSet": [
					{"jsonPath": "$.services.*.description"},
					{"jsonPath": "$.services", "processingOptions": {"actions":[ "create", "delete"]}}
				]
			},
			{
				"roleIds": ["developer"],
				"disallowedRuleSet": [
					{"ruleId": "endpoint.security.edit"}
				],
				"isInheritedFromTenant": true
			},
			{
				"roleIds": ["some-role"],
				"allowedRuleSet": [
          { "jsonPath": "$.endpoints.*.public" },
        	{ "jsonPath": "$.secrets", "processingOptions": { "actions": ["create", "delete"], "primaryKey": "clientType" }}
				]
			}
		]
	}
}`
	validBodyStringBeforeV14_1 := `{
	"tenantId": "company-1",
	"configurationManagement": {
		"saveChangesRules": [
			{
				"roleIds": ["maintainer"],
				"disallowedRuleSet": [
					{"jsonPath": "$.services.*.description"},
					{"jsonPath": "$.services", "processingOptions": {"action": "create"}}
				]
			},
			{
				"roleIds": ["developer"],
				"disallowedRuleSet": [
					{"ruleId": "endpoint.security.edit"}
				],
				"isInheritedFromTenant": true
			}
		]
	}
}`

	expectedRes := []*rulesentities.ProjectSaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
			IsInheritedFromTenant: true,
		},
		{
			RoleIDs: []string{"some-role"},
			AllowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.endpoints.*.public"},
				{JSONPath: "$.secrets", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}, PrimaryKey: "clientType"}},
			},
		},
	}

	expecteResBeforeV14_1 := []*rulesentities.ProjectSaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Action: "create"}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
			IsInheritedFromTenant: true,
		},
	}

	testCases := map[string]struct {
		companyID      string
		server         *httptest.Server
		err            bool
		expectedResult []*rulesentities.ProjectSaveChangesRules
	}{
		"valid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/projects/" + "my-project",
				verb: http.MethodGet,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   validBodyString,
			}),
			expectedResult: expectedRes,
		},
		"valid response - before console V14.1": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/projects/" + "my-project",
				verb: http.MethodGet,
			}, MockResponse{
				statusCode: http.StatusOK,
				respBody:   validBodyStringBeforeV14_1,
			}),
			expectedResult: expecteResBeforeV14_1,
		},
		"invalid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: "/api/backend/projects/" + "my-project",
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

			data, err := New(client).ListProjectRules(t.Context(), testCase.companyID)
			if testCase.err {
				require.Error(t, err)
				require.Nil(t, data)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, data)
			}
		})
	}
}

func TestClientTenantPatch(t *testing.T) {
	patchData := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
		{
			RoleIDs: []string{"some-role"},
			AllowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.endpoints.*.public"},
				{JSONPath: "$.secrets", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}, PrimaryKey: "clientType"}},
			},
		},
	}

	patchDataBeforeV14_1 := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Action: "create"}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
	}

	testCases := map[string]struct {
		companyID string
		PatchData []*rulesentities.SaveChangesRules
		server    *httptest.Server
		err       bool
	}{
		"valid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path:              fmt.Sprintf("/api/backend/tenants/%s/rules", "company-1"),
				verb:              http.MethodPatch,
				bodyContainsMatch: `"actions":["create","delete"]`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
			PatchData: patchData,
		},
		"valid response - before Console V14.1": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path:              fmt.Sprintf("/api/backend/tenants/%s/rules", "company-1"),
				verb:              http.MethodPatch,
				bodyContainsMatch: `"action":"create"`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
			PatchData: patchDataBeforeV14_1,
		},
		"invalid response": {
			companyID: "company-1",
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf("/api/backend/tenants/%s/rules", "company-1"),
				verb: http.MethodPatch,
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

			err = New(client).UpdateTenantRules(t.Context(), testCase.companyID, testCase.PatchData)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientProjectPatch(t *testing.T) {
	patchData := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
		{
			RoleIDs: []string{"some-role"},
			AllowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.endpoints.*.public"},
				{JSONPath: "$.secrets", Options: &rulesentities.RuleOptions{Actions: []string{"create", "delete"}, PrimaryKey: "clientType"}},
			},
		},
	}

	patchDataBeforeV14_1 := []*rulesentities.SaveChangesRules{
		{
			RoleIDs: []string{"maintainer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{JSONPath: "$.services.*.description"},
				{JSONPath: "$.services", Options: &rulesentities.RuleOptions{Action: "create"}},
			},
		},
		{
			RoleIDs: []string{"developer"},
			DisallowedRuleSet: []rulesentities.RuleSet{
				{RuleID: "endpoint.security.edit"},
			},
		},
	}

	testCases := map[string]struct {
		projectID string
		PatchData []*rulesentities.SaveChangesRules
		server    *httptest.Server
		err       bool
	}{
		"valid response": {
			projectID: "project-1",
			server: mockServer(t, ExpectedRequest{
				path:              fmt.Sprintf("/api/backend/projects/%s/rules", "project-1"),
				verb:              http.MethodPatch,
				bodyContainsMatch: `"actions":["create","delete"]`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
			PatchData: patchData,
		},
		"valid response - before Console V14.1": {
			projectID: "project-1",
			server: mockServer(t, ExpectedRequest{
				path:              fmt.Sprintf("/api/backend/projects/%s/rules", "project-1"),
				verb:              http.MethodPatch,
				bodyContainsMatch: `"action":"create"`,
			}, MockResponse{
				statusCode: http.StatusOK,
			}),
			PatchData: patchDataBeforeV14_1,
		},
		"invalid response": {
			projectID: "project-1",
			server: mockServer(t, ExpectedRequest{
				path: fmt.Sprintf("/api/backend/projects/%s/rules", "project-1"),
				verb: http.MethodPatch,
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

			err = New(client).UpdateTenantRules(t.Context(), testCase.projectID, testCase.PatchData)
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
	path              string
	verb              string
	bodyContainsMatch string
}

func mockServer(t *testing.T, expectedReq ExpectedRequest, resp MockResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.RequestURI != expectedReq.path && req.Method != expectedReq.verb {
			w.WriteHeader(http.StatusNotFound)
			assert.Fail(t, fmt.Sprintf("unsupported call: %s - wanted: %s", req.RequestURI, expectedReq.path))
			return
		}

		if expectedReq.bodyContainsMatch != "" {
			foundBody, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Contains(t, string(foundBody), expectedReq.bodyContainsMatch)
		}

		w.WriteHeader(resp.statusCode)
		if resp.err {
			w.Write([]byte(`{"error":"some error","message":"some message"}`))
			return
		}
		w.Write([]byte(resp.respBody))
	}))
}
