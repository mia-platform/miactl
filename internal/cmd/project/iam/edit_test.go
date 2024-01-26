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
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditRoleForEntity(t *testing.T) {
	companyID := "company"
	projectID := "project"
	entityID := "entity-id"
	role := string(resources.IAMRoleDeveloper)
	wrongRole := "wrong"
	testCases := map[string]struct {
		server     *httptest.Server
		roleChange roleChanges
		err        bool
	}{
		"update user": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   projectID,
				entityID:    entityID,
				entityType:  iam.UsersEntityName,
				projectRole: &role,
			},
		},
		"update group": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.GroupsEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   projectID,
				entityID:    entityID,
				entityType:  iam.GroupsEntityName,
				projectRole: &role,
			},
		},
		"update service account": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   projectID,
				entityID:    entityID,
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: &role,
			},
		},
		"missing company ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:   "",
				projectID:   projectID,
				entityID:    entityID,
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: &role,
			},

			err: true,
		},
		"missing project ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   "",
				entityID:    entityID,
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: &role,
			},
			err: true,
		},
		"missing entity ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   projectID,
				entityID:    "",
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: &role,
			},
			err: true,
		},
		"invalid project role": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:   companyID,
				projectID:   projectID,
				entityID:    entityID,
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: &wrongRole,
			},
			err: true,
		},
		"invalid environment role": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:       companyID,
				projectID:       projectID,
				entityID:        entityID,
				entityType:      iam.ServiceAccountsEntityName,
				projectRole:     &role,
				environmentName: "environment",
				environmentRole: wrongRole,
			},
			err: true,
		},
		"invalid body response": {
			server: iam.ErrorTestServerForEditIAMRole(t, companyID, entityID),
			err:    true,
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
			err = editRoleForEntity(context.TODO(), client, testCase.roleChange)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
