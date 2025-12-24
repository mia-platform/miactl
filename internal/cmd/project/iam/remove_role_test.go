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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/iam"
)

func TestRemoveRoleForEntity(t *testing.T) {
	companyID := "company"
	projectID := "project"
	entityID := "entity-id"
	testCases := map[string]struct {
		server     *httptest.Server
		roleChange roleChanges
		err        bool
	}{
		"remove role to user": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			roleChange: roleChanges{
				companyID:  companyID,
				projectID:  projectID,
				entityID:   entityID,
				entityType: iam.UsersEntityName,
			},
		},
		"remove role to group": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.GroupsEntityName),
			roleChange: roleChanges{
				companyID:  companyID,
				projectID:  projectID,
				entityID:   entityID,
				entityType: iam.GroupsEntityName,
			},
		},
		"remove role to service account": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:  companyID,
				projectID:  projectID,
				entityID:   entityID,
				entityType: iam.ServiceAccountsEntityName,
			},
		},
		"missing company ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:  "",
				projectID:  projectID,
				entityID:   entityID,
				entityType: iam.ServiceAccountsEntityName,
			},

			err: true,
		},
		"missing project ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:  companyID,
				projectID:  "",
				entityID:   entityID,
				entityType: iam.ServiceAccountsEntityName,
			},
			err: true,
		},
		"missing entity ID": {
			server: iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.ServiceAccountsEntityName),
			roleChange: roleChanges{
				companyID:  companyID,
				projectID:  projectID,
				entityID:   "",
				entityType: iam.ServiceAccountsEntityName,
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
			err = removeRoleForEntity(t.Context(), client, testCase.roleChange)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
