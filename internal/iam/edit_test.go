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
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditRoleForIdentity(t *testing.T) {
	companyID := "company-id"
	entityID := "entity-id"

	testCases := map[string]struct {
		server    *httptest.Server
		iamType   string
		iamRole   resources.EditIAMRole
		expectErr bool
	}{
		"edit user": {
			server:  TestServerForCompanyIAMEditRole(t, companyID, entityID, UsersEntityName),
			iamType: UsersEntityName,
			iamRole: resources.EditIAMRole{
				Role: resources.IAMRoleCompanyOwner,
			},
		},
		"edit group": {
			server:  TestServerForCompanyIAMEditRole(t, companyID, entityID, GroupsEntityName),
			iamType: GroupsEntityName,
			iamRole: resources.EditIAMRole{
				Role: resources.IAMRoleCompanyOwner,
			},
		},
		"edit service account": {
			server:  TestServerForCompanyIAMEditRole(t, companyID, entityID, ServiceAccountsEntityName),
			iamType: ServiceAccountsEntityName,
			iamRole: resources.EditIAMRole{
				Role: resources.IAMRoleCompanyOwner,
			},
		},
		"handle error": {
			server:  ErrorTestServerForEditIAMRole(t, companyID, entityID),
			iamType: UsersEntityName,
			iamRole: resources.EditIAMRole{
				Role: resources.IAMRoleCompanyOwner,
			},
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

			response, err := EditIAMResourceRole(context.TODO(), client, companyID, entityID, testCase.iamType, testCase.iamRole)
			require.NoError(t, err)
			if testCase.expectErr {
				assert.Error(t, response.Error())
				return
			}

			assert.NoError(t, response.Error())
		})
	}
}
