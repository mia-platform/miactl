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

package user

import (
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditUser(t *testing.T) {
	companyID := "company-id"
	entityID := "000000000000000000000001"
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		role      resources.IAMRole
		userID    string
		expectErr bool
	}{
		"edit user": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			companyID: companyID,
			role:      resources.IAMRoleGuest,
			userID:    entityID,
		},
		"missing company": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			companyID: "",
			role:      resources.IAMRoleGuest,
			userID:    entityID,
			expectErr: true,
		},
		"missing user id": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			companyID: companyID,
			role:      resources.IAMRoleGuest,
			userID:    "",
			expectErr: true,
		},
		"wrong role": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, entityID, iam.UsersEntityName),
			companyID: "",
			role:      resources.IAMRole("example"),
			userID:    entityID,
			expectErr: true,
		},
		"error from backend": {
			server:    iam.ErrorTestServerForEditIAMRole(t, companyID, entityID),
			companyID: companyID,
			role:      resources.IAMRoleCompanyOwner,
			userID:    entityID,
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.server
			defer server.Close()
			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)
			err = editCompanyUser(
				t.Context(),
				client,
				testCase.companyID,
				testCase.userID,
				testCase.role,
			)

			switch testCase.expectErr {
			case true:
				assert.Error(t, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}
