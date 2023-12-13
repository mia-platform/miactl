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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllIAMIdentities(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companyID    string
		searchParams map[string]bool
		err          bool
	}{
		"valid get response": {
			server:    mockListServer(t),
			companyID: "success",
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			searchParams: map[string]bool{},
		},
		"valid get with search parameters": {
			server:    mockListServer(t),
			companyID: "search",
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			searchParams: map[string]bool{
				ServiceAccountsEntityName: true,
				GroupsEntityName:          true,
			},
		},
		"invalid body response": {
			server:    mockListServer(t),
			companyID: "fail",
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)
			err = listAllIAMEntities(context.TODO(), client, testCase.companyID, testCase.searchParams)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListUsersIdentities(t *testing.T) {
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companyID    string
		searchParams map[string]bool
		err          bool
	}{
		"valid get response": {
			server:    mockListServer(t),
			companyID: "success",
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			searchParams: map[string]bool{},
		},
		"invalid body response": {
			server:    mockListServer(t),
			companyID: "fail",
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)

			err = listSpecificEntities(context.TODO(), client, testCase.companyID, UsersEntityName)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockListServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, err := url.ParseQuery(r.URL.RawQuery)
		require.NoError(t, err)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, "success"):
			assert.Equal(t, 0, len(params))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validListIAMIdentitiesBodyString))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, "search"):
			searchTerms, ok := params["identityType"]
			assert.True(t, ok)
			assert.ElementsMatch(t, []string{"group", "serviceAccount"}, searchTerms)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(filteredListIAMIdentitiesString))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listUserEntityTemplate, "success"):
			assert.Equal(t, 0, len(params))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validListUserIdentitiesBodyString))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, "fail"):
			assert.Equal(t, 0, len(params))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listUserEntityTemplate, "fail"):
			assert.Equal(t, 0, len(params))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		default:
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call")
		}
	}))
}

const (
	validListIAMIdentitiesBodyString = `[
  {
    "identityId": "000000000000000000000000",
    "email": "user.email@example.com",
    "name": "User Complete Name",
    "identityType": "user",
    "companyRoles": ["guest"],
    "lastLogin": "0001-01-01T00:00:00.000Z"
  },
  {
    "identityId": "000000000000000000000001",
    "name": "Group Name",
    "identityType": "group",
    "companyRoles": ["developer"],
    "lastLogin": "0001-01-01T00:00:00.000Z",
    "membersCount": 1,
    "members": [
      {
        "name": "User Complete Name",
        "email": "user.email@example.com"
      }
    ]
  },
  {
    "identityId": "000000000000000000000002",
    "name": "Service Account Name",
    "identityType": "serviceAccount",
    "companyRoles": ["project-admin"],
    "lastLogin": "0001-01-01T00:00:00.000Z",
    "authMethod": "client_secret_basic"
  }
]`
	filteredListIAMIdentitiesString = `[
    {
      "identityId": "000000000000000000000001",
      "name": "Group Name",
      "identityType": "group",
      "companyRoles": ["developer"],
      "lastLogin": "0001-01-01T00:00:00.000Z",
      "membersCount": 1,
      "members": [
        {
          "name": "User Complete Name",
          "email": "user.email@example.com"
        }
      ]
    },
    {
      "identityId": "000000000000000000000002",
      "name": "Service Account Name",
      "identityType": "serviceAccount",
      "companyRoles": ["project-admin"],
      "lastLogin": "0001-01-01T00:00:00.000Z",
      "authMethod": "client_secret_basic"
    }
]`
	validListUserIdentitiesBodyString = `[
  {
    "userId": "000000000000000000000001",
    "email": "user.email@example.com",
    "fullName": "User Full Name",
    "companyRoles": [],
    "lastLogin": "2010-01-01T00:00:00.000Z",
    "groups": [{
      "name": "Role Name",
      "roleId": "role-id"
    }]
  },
  {
    "userId": "000000000000000000000002",
    "email": "user.email@example.com",
    "fullName": "User Full Name",
    "companyRoles": ["role-id"],
    "lastLogin": "2010-01-01T00:00:00.000Z"
  },
  {
    "userId": "000000000000000000000003",
    "email": "user.email@example.com",
    "fullName": "User Full Name",
    "companyRoles": ["role-id"],
    "lastLogin": "2010-01-01T00:00:00.000Z",
    "groups": [{
      "name": "Role Name",
      "roleId": "role-id"
    }]
  }
]`
)
