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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroup(t *testing.T) {
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		role      resources.ServiceAccountRole
		groupName string
		expectErr bool
	}{
		"create group": {
			server:    addUserTestServer(t),
			companyID: "success",
			role:      resources.ServiceAccountRoleGuest,
			groupName: "group-name",
		},
		"missing company": {
			server:    addUserTestServer(t),
			companyID: "",
			role:      resources.ServiceAccountRoleGuest,
			groupName: "group-name",
			expectErr: true,
		},
		"missing group name": {
			server:    addUserTestServer(t),
			companyID: "success",
			role:      resources.ServiceAccountRoleGuest,
			groupName: "",
			expectErr: true,
		},
		"wrong role": {
			server:    addUserTestServer(t),
			companyID: "succes",
			role:      resources.ServiceAccountRole("example"),
			groupName: "group-name",
			expectErr: true,
		},
		"error from backend": {
			server:    addUserTestServer(t),
			companyID: "fail",
			role:      resources.ServiceAccountRoleCompanyOwner,
			groupName: "group-name",
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
			err = createNewGroup(
				context.TODO(),
				client,
				testCase.companyID,
				testCase.groupName,
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

func addUserTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createGroupTemplate, "success"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createGroupTemplate, "fail"):
			w.WriteHeader(http.StatusBadRequest)
		default:
			require.Fail(t, "request not implemented", "request received for %s with %s method", r.URL, r.Method)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
