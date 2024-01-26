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

package iam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllIAMEntitites(t *testing.T) {
	companyID := "company-id"
	projectID := "project-id"
	testCases := map[string]struct {
		server     *httptest.Server
		projectIds []string
		expectErr  bool
	}{
		"list all entities in company": {
			server: TestServerForCompanyIAMList(t, companyID),
		},
		"list all entities in project": {
			server:     TestServerForProjectIAMList(t, companyID, projectID),
			projectIds: []string{projectID},
		},
		"test error for company": {
			server:    ErrorTestServerForCompanyIAMList(t, companyID),
			expectErr: true,
		},
		"test error for project": {
			server:     ErrorTestServerForProjectIAMList(t, companyID, projectID),
			projectIds: []string{projectID},
			expectErr:  true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			config := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}
			client, err := client.APIClientForConfig(config)
			require.NoError(t, err)

			response, err := ListAllIAMEntities(context.TODO(), client, companyID, testCase.projectIds, nil)
			require.NoError(t, err)
			if testCase.expectErr {
				assert.Error(t, response.Error())
				return
			}

			assert.NoError(t, response.Error())
		})
	}
}

func TestSpecificIAMList(t *testing.T) {
	companyID := "company-id"
	testCases := map[string]struct {
		server    *httptest.Server
		iamType   string
		expectErr bool
	}{
		"list users in company": {
			server:  TestServerForCompanySpecificList(t, companyID, UsersEntityName),
			iamType: UsersEntityName,
		},
		"list groups in company": {
			server:  TestServerForCompanySpecificList(t, companyID, GroupsEntityName),
			iamType: GroupsEntityName,
		},
		"list service accounts in company": {
			server:  TestServerForCompanySpecificList(t, companyID, ServiceAccountsEntityName),
			iamType: ServiceAccountsEntityName,
		},
		"test error for company": {
			server:    ErrorTestServerForCompanyIAMList(t, companyID),
			iamType:   UsersEntityName,
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			config := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}
			client, err := client.APIClientForConfig(config)
			require.NoError(t, err)

			response, err := ListSpecificEntities(context.TODO(), client, companyID, testCase.iamType)
			require.NoError(t, err)
			if testCase.expectErr {
				assert.Error(t, response.Error())
				return
			}

			assert.NoError(t, response.Error())
		})
	}
}
