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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddGroupMember(t *testing.T) {
	testCases := map[string]struct {
		server     *httptest.Server
		companyID  string
		groupID    string
		userEmails []string
		expectErr  bool
	}{
		"add member to group": {
			server:     addGroupMemeberTestServer(t),
			companyID:  "success",
			groupID:    "group-id",
			userEmails: []string{"user@example.com"},
		},
		"missing company": {
			server:     addGroupMemeberTestServer(t),
			companyID:  "",
			groupID:    "group-id",
			userEmails: []string{"user@example.com"},
			expectErr:  true,
		},
		"missing group id": {
			server:     addGroupMemeberTestServer(t),
			companyID:  "success",
			groupID:    "",
			userEmails: []string{"user@example.com"},
			expectErr:  true,
		},
		"missing user email": {
			server:     addGroupMemeberTestServer(t),
			companyID:  "succes",
			groupID:    "group-id",
			userEmails: []string{},
			expectErr:  true,
		},
		"error from backend": {
			server:     addGroupMemeberTestServer(t),
			companyID:  "fail",
			groupID:    "group-id",
			userEmails: []string{"user@example.com"},
			expectErr:  true,
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
			err = addMemberToGroup(
				t.Context(),
				client,
				testCase.companyID,
				testCase.groupID,
				testCase.userEmails,
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

func addGroupMemeberTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(addMemberTemplate, "success", "group-id"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(addMemberTemplate, "fail", "group-id"):
			w.WriteHeader(http.StatusBadRequest)
		default:
			assert.Fail(t, "request not implemented", "request received for %s with %s method", r.URL, r.Method)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
