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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveGroupMember(t *testing.T) {
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		groupID   string
		userIds   []string
		expectErr bool
	}{
		"remove member to group": {
			server:    removeGroupMemeberTestServer(t),
			companyID: "success",
			groupID:   "group-id",
			userIds:   []string{"000000000000000000000001"},
		},
		"missing company": {
			server:    removeGroupMemeberTestServer(t),
			companyID: "",
			groupID:   "group-id",
			userIds:   []string{"000000000000000000000001"},
			expectErr: true,
		},
		"missing group id": {
			server:    removeGroupMemeberTestServer(t),
			companyID: "success",
			groupID:   "",
			userIds:   []string{"000000000000000000000001"},
			expectErr: true,
		},
		"missing user id": {
			server:    removeGroupMemeberTestServer(t),
			companyID: "succes",
			groupID:   "group-id",
			userIds:   []string{},
			expectErr: true,
		},
		"error from backend": {
			server:    removeGroupMemeberTestServer(t),
			companyID: "fail",
			groupID:   "group-id",
			userIds:   []string{"000000000000000000000001"},
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
			err = removeMemberFromGroup(
				context.TODO(),
				client,
				testCase.companyID,
				testCase.groupID,
				testCase.userIds,
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

func removeGroupMemeberTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf(removeMemberTemplate, "success", "group-id"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf(removeMemberTemplate, "fail", "group-id"):
			w.WriteHeader(http.StatusBadRequest)
		default:
			require.Fail(t, "request not implemented", "request received for %s with %s method", r.URL, r.Method)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
