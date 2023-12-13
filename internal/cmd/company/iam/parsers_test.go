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
				Type:  "group",
				Roles: []string{"company-owner", "guest"},
			},
			expectedRow: []string{"Group", "identity name", "Company Owner, Guest"},
		},
		"user account IAM identity": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  "serviceAccount",
				Roles: []string{"developer"},
			},
			expectedRow: []string{"Service Account", "identity name", "Developer"},
		},
		"user IAM identity": {
			identity: resources.IAMIdentity{
				ID:    "identity-id",
				Name:  "identity name",
				Type:  "user",
				Roles: []string{"developer"},
			},
			expectedRow: []string{"User", "identity name", "Developer"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForIAMIdentity(testCase.identity))
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
			expectedRow: []string{"identity name", "user@example.com", "Company Owner, Guest", "-", "0s"},
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
			expectedRow: []string{"identity name", "user@example.com", "-", "group name", "0s"},
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
			expectedRow: []string{"identity name", "user@example.com", "Company Owner", "group name, group name2", "0s"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForUserIdentity(testCase.identity))
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
				Name: "identity name",
				Role: "guest",
			},
			expectedRow: []string{"identity name", "Guest", "-"},
		},
		"group without one user": {
			identity: resources.GroupIdentity{
				Name: "identity name",
				Role: "project-admin",
				Members: []resources.UserIdentity{
					{
						Name: "user name",
					},
				},
			},
			expectedRow: []string{"identity name", "Project Administrator", "user name"},
		},
		"group without more users": {
			identity: resources.GroupIdentity{
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
			expectedRow: []string{"identity name", "Guest", "user name, name user"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForGroupIdentity(testCase.identity))
		})
	}
}
