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
	testCases := map[string]struct {
		server     *httptest.Server
		roleChange roleChanges
		err        bool
	}{
		"update user": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			roleChange: roleChanges{
				entityID:    entityID,
				entityType:  iam.UsersEntityName,
				projectRole: resources.IAMRoleDeveloper,
			},
		},
		"update group": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.GroupsEntityName),
			roleChange: roleChanges{
				entityID:    entityID,
				entityType:  iam.GroupsEntityName,
				projectRole: resources.IAMRoleDeveloper,
			},
		},
		"update service account": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				entityID:    entityID,
				entityType:  iam.ServiceAccountsEntityName,
				projectRole: resources.IAMRoleDeveloper,
			},
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
			err = editRoleForEntity(context.TODO(), client, companyID, projectID, testCase.roleChange)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPayloadForChanges(t *testing.T) {
	projectID := "project-id"
	environmentID := "environment-id"
	projectRole := resources.IAMRoleProjectAdmin
	environmenRole := resources.IAMRoleGuest
	testCases := map[string]struct {
		expectedPayload resources.EditIAMRole
		changes         roleChanges
	}{
		"change only project role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      projectRole,
					},
				},
			},
			changes: roleChanges{
				projectRole: projectRole,
			},
		},
		"change only environment role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          environmenRole,
							},
						},
					},
				},
			},
			changes: roleChanges{
				environmentName: environmentID,
				environmentRole: environmenRole,
			},
		},
		"both roles": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      projectRole,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          environmenRole,
							},
						},
					},
				},
			},
			changes: roleChanges{
				projectRole:     projectRole,
				environmentName: environmentID,
				environmentRole: environmenRole,
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			payload := payloadForChanges(projectID, testCase.changes)
			assert.Equal(t, testCase.expectedPayload, payload)
		})
	}
}
