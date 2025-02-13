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

package group

import (
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditGroup(t *testing.T) {
	companyID := "company-id"
	groupID := "000000000000000000000001"
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		role      resources.IAMRole
		groupID   string
		expectErr bool
	}{
		"edit group": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, groupID, iam.GroupsEntityName),
			companyID: companyID,
			role:      resources.IAMRoleGuest,
			groupID:   groupID,
		},
		"missing company": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, groupID, iam.GroupsEntityName),
			companyID: "",
			role:      resources.IAMRoleGuest,
			groupID:   groupID,
			expectErr: true,
		},
		"missing group id": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, groupID, iam.GroupsEntityName),
			companyID: companyID,
			role:      resources.IAMRoleGuest,
			groupID:   "",
			expectErr: true,
		},
		"wrong role": {
			server:    iam.TestServerForCompanyIAMEditRole(t, companyID, groupID, iam.GroupsEntityName),
			companyID: "",
			role:      resources.IAMRole("example"),
			groupID:   groupID,
			expectErr: true,
		},
		"error from backend": {
			server:    iam.ErrorTestServerForEditIAMRole(t, companyID, groupID),
			companyID: companyID,
			role:      resources.IAMRoleCompanyOwner,
			groupID:   groupID,
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
			err = editCompanyGroup(
				t.Context(),
				client,
				testCase.companyID,
				testCase.groupID,
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
