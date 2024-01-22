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
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestRowForIAMIdentity(t *testing.T) {
	testCases := map[string]struct {
		identity    resources.IAMIdentity
		expectedRow []string
	}{
		"group IAM identity": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  GroupsEntityName,
				Roles: []string{"company-owner", "guest"},
			},
			expectedRow: []string{"identity-id", "Group", "identity name", "Company Owner, Guest"},
		},
		"user account IAM identity": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  ServiceAccountsEntityName,
				Roles: []string{"developer"},
			},
			expectedRow: []string{"identity-id", "Service Account", "identity name", "Developer"},
		},
		"user IAM identity": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  UsersEntityName,
				Roles: []string{"developer"},
			},
			expectedRow: []string{"identity-id", "User", "identity name", "Developer"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, RowForIAMIdentity(testCase.identity))
		})
	}
}

func TestRowForProjectIAMIdentity(t *testing.T) {
	projectID := "000000000000000000000001"
	testCases := map[string]struct {
		identity    resources.IAMIdentity
		expectedRow []string
	}{
		"IAM identity without project roles": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  GroupsEntityName,
				Roles: []string{"company-owner", "guest"},
			},
			expectedRow: []string{"identity-id", "Group", "identity name", "Company Owner (inherited)", ""},
		},
		"IAM identity with project role": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  ServiceAccountsEntityName,
				Roles: []string{"developer"},
				ProjectsRole: []resources.ProjectRole{
					{
						ID:    projectID,
						Roles: []string{"guest"},
					},
				},
			},
			expectedRow: []string{"identity-id", "Service Account", "identity name", "Guest", ""},
		},
		"IAM identity with empty project role": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  UsersEntityName,
				Roles: []string{"developer"},
				ProjectsRole: []resources.ProjectRole{
					{
						ID:    projectID,
						Roles: []string{},
					},
				},
			},
			expectedRow: []string{"identity-id", "User", "identity name", "Developer (inherited)", ""},
		},
		"IAM with other projects access": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  UsersEntityName,
				Roles: []string{"developer"},
				ProjectsRole: []resources.ProjectRole{
					{
						ID:    "other-id",
						Roles: []string{"guest"},
					},
				},
			},
			expectedRow: []string{"identity-id", "User", "identity name", "Developer (inherited)", ""},
		},
		"IAM with environment specific access": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  UsersEntityName,
				Roles: []string{"developer"},
				ProjectsRole: []resources.ProjectRole{
					{
						ID:    "other-id",
						Roles: []string{"guest"},
						Environments: []resources.EnvironmentRole{
							{
								ID:    "envId",
								Roles: []string{"developer"},
							},
						},
					},
				},
			},
			expectedRow: []string{"identity-id", "User", "identity name", "Developer (inherited)", "envId=Developer"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, RowForProjectIAMIdentity(projectID)(testCase.identity))
		})
	}
}

func TestRowForUserIdentity(t *testing.T) {
	testCases := map[string]struct {
		identity    resources.UserIdentity
		expectedRow []string
	}{
		"user with role": {
			identity: resources.UserIdentity{
				ID:        "identity-id",
				Email:     "user@example.com",
				FullName:  "identity name",
				Roles:     []string{"company-owner", "guest"},
				LastLogin: time.Now(),
			},
			expectedRow: []string{"identity-id", "identity name", "user@example.com", "Company Owner, Guest", "-", "0s"},
		},
		"user with groups": {
			identity: resources.UserIdentity{
				ID:       "identity-id",
				Email:    "user@example.com",
				FullName: "identity name",
				Groups: []resources.GroupIdentity{
					{
						Name: "group name",
						Role: "guest",
					},
				},
				LastLogin: time.Now(),
			},
			expectedRow: []string{"identity-id", "identity name", "user@example.com", "-", "group name", "0s"},
		},
		"user with both": {
			identity: resources.UserIdentity{
				ID:       "identity-id",
				Email:    "user@example.com",
				FullName: "identity name",
				Roles:    []string{"company-owner"},
				Groups: []resources.GroupIdentity{
					{
						Name: "group name",
						Role: "guest",
					},
					{
						Name: "group name2",
						Role: "guest",
					},
				},
				LastLogin: time.Now(),
			},
			expectedRow: []string{"identity-id", "identity name", "user@example.com", "Company Owner", "group name, group name2", "0s"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, RowForUserIdentity(testCase.identity))
		})
	}
}

func TestRowForGroupIdentity(t *testing.T) {
	testCases := map[string]struct {
		identity    resources.GroupIdentity
		expectedRow []string
	}{
		"group without users": {
			identity: resources.GroupIdentity{
				ID:   "identity-id",
				Name: "identity name",
				Role: "guest",
			},
			expectedRow: []string{"identity-id", "identity name", "Guest", "-"},
		},
		"group without one user": {
			identity: resources.GroupIdentity{
				ID:   "identity-id",
				Name: "identity name",
				Role: "project-admin",
				Members: []resources.UserIdentity{
					{
						Name: "user name",
					},
				},
			},
			expectedRow: []string{"identity-id", "identity name", "Project Administrator", "user name"},
		},
		"group without more users": {
			identity: resources.GroupIdentity{
				ID:   "identity-id",
				Name: "identity name",
				Role: "guest",
				Members: []resources.UserIdentity{
					{
						Name: "user name",
					},
					{
						Name: "name user",
					},
				},
			},
			expectedRow: []string{"identity-id", "identity name", "Guest", "user name, name user"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, RowForGroupIdentity(testCase.identity))
		})
	}
}

func TestRowForServiceAccountIdentity(t *testing.T) {
	testCases := map[string]struct {
		identity    resources.ServiceAccountIdentity
		expectedRow []string
	}{
		"base service account": {
			identity: resources.ServiceAccountIdentity{
				ID:        "identity-id",
				Name:      "identity name",
				Roles:     []string{"guest"},
				LastLogin: time.Now(),
			},
			expectedRow: []string{"identity-id", "identity name", "Guest", "0s"},
		},
		"service account without login and with multiple roles": {
			identity: resources.ServiceAccountIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Roles: []string{"guest, developer"},
			},
			expectedRow: []string{"identity-id", "identity name", "Guest, Developer", "-"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, RowForServiceAccountIdentity(testCase.identity))
		})
	}
}
