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
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ErrorTestServerForCompanyIAM(t *testing.T, companyID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		require.False(t, params.Has(projectIdsKey), "for company calls no project ids are allowed")
		internalErrorServerHandler(t, w, r, companyID)
	}))
}

func ErrorTestServerForProjectIAM(t *testing.T, companyID, projectID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		require.True(t, params.Has(projectIdsKey), "for project calls project ids are required")
		projectIDs := params[projectIdsKey]
		require.Equal(t, 1, len(projectIDs), "only one project id is required")
		assert.Equal(t, projectID, projectIDs[0])
		internalErrorServerHandler(t, w, r, companyID)
	}))
}

func internalErrorServerHandler(t *testing.T, w http.ResponseWriter, r *http.Request, companyID string) {
	t.Helper()
	switch {
	case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, companyID):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listUsersEntityTemplate, companyID):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listGroupsEntityTemplate, companyID):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listServiceAccountsEntityTemplate, companyID):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	default:
		w.WriteHeader(http.StatusNotFound)
		require.Fail(t, "unsupported call", "%q, %q", r.Method, r.URL.String())
	}
}

func TestServerForCompanyIAMList(t *testing.T, companyID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		require.False(t, params.Has(projectIdsKey), "for company calls no project ids are allowed")
		searchTerms := params["identityType"]
		if len(searchTerms) == 0 {
			searchTerms = append(searchTerms, UsersEntityName, GroupsEntityName, ServiceAccountsEntityName)
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, companyID):
			w.WriteHeader(http.StatusOK)

			var identities []interface{}
			if slices.Contains(searchTerms, UsersEntityName) {
				identities = append(identities, userExample)
			}
			if slices.Contains(searchTerms, GroupsEntityName) {
				identities = append(identities, groupExample)
			}
			if slices.Contains(searchTerms, ServiceAccountsEntityName) {
				identities = append(identities, serviceAccountExample)
			}

			payload, err := resources.EncodeResourceToJSON(identities)
			require.NoError(t, err)
			_, _ = w.Write(payload)
		default:
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call", "%q, %q", r.Method, r.URL.String())
		}
	}))
}

func TestServerForProjectIAMList(t *testing.T, companyID, projectID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		require.True(t, params.Has(projectIdsKey), "for project calls project ids are required")
		projectIDs := params[projectIdsKey]
		require.Equal(t, 1, len(projectIDs), "only one project id is required")
		assert.Equal(t, projectID, projectIDs[0])

		searchTerms := params["identityType"]
		if len(searchTerms) == 0 {
			searchTerms = append(searchTerms, UsersEntityName, GroupsEntityName, ServiceAccountsEntityName)
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(listAllIAMEntitiesTemplate, companyID):
			w.WriteHeader(http.StatusOK)

			var identities []interface{}
			if slices.Contains(searchTerms, UsersEntityName) {
				identities = append(identities, userExample)
			}
			if slices.Contains(searchTerms, GroupsEntityName) {
				identities = append(identities, groupExample)
			}
			if slices.Contains(searchTerms, ServiceAccountsEntityName) {
				identities = append(identities, serviceAccountExample)
			}

			payload, err := resources.EncodeResourceToJSON(identities)
			require.NoError(t, err)
			_, _ = w.Write(payload)
		default:
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call", "%q, %q", r.Method, r.URL.String())
		}
	}))
}

func TestServerForCompanySpecificList(t *testing.T, companyID string, entityType string) *httptest.Server {
	t.Helper()
	var pathTemplate string
	var responseResources []interface{}
	switch entityType {
	case UsersEntityName:
		pathTemplate = listUsersEntityTemplate
		responseResources = []interface{}{userExample}
	case GroupsEntityName:
		pathTemplate = listGroupsEntityTemplate
		responseResources = []interface{}{groupExample}
	case ServiceAccountsEntityName:
		pathTemplate = listServiceAccountsEntityTemplate
		responseResources = []interface{}{serviceAccountExample}
	default:
		require.FailNow(t, "unrecognized entity type", "%q", entityType)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		require.Equal(t, 0, len(params), "no query param supported")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(pathTemplate, companyID):
			w.WriteHeader(http.StatusOK)
			payload, err := resources.EncodeResourceToJSON(responseResources)
			require.NoError(t, err)
			_, _ = w.Write(payload)
		default:
			w.WriteHeader(http.StatusNotFound)
			require.Fail(t, "unsupported call", "%q, %q", r.Method, r.URL.String())
		}
	}))
}

func testTime() time.Time {
	result, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00.000Z")
	return result
}

var (
	userExample = &resources.UserIdentity{
		ID:    "000000000000000000000000",
		Email: "user@example.com",
		Name:  "User Fulle Name",
		Roles: []string{
			string(resources.IAMRoleCompanyOwner),
		},
		LastLogin: testTime(),
		Groups: []resources.GroupIdentity{
			{
				Name:   "Group Name",
				RoleID: string(resources.IAMRoleProjectAdmin),
			},
			{
				Name:   "Second Group Name",
				RoleID: string(resources.IAMRoleGuest),
			},
		},
	}
	groupExample = &resources.GroupIdentity{
		ID:   "000000000000000000000001",
		Name: "",
		Role: string(resources.IAMRoleCompanyOwner),
		Members: []resources.UserIdentity{
			{
				Name: "User Full Name",
			},
			{
				Name: "Other User Full Name",
			},
		},
	}
	serviceAccountExample = &resources.ServiceAccountIdentity{
		ID:         "000000000000000000000001",
		Name:       "service account name",
		AuthMethod: "client_secret_basic",
		Roles: []string{
			string(resources.IAMRoleReporter),
		},
		LastLogin: testTime(),
	}
)
