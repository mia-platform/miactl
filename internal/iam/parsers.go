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
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/util"
)

func RowForIAMIdentity(identity resources.IAMIdentity) []string {
	caser := cases.Title(language.English)
	return []string{
		identity.ID,
		caser.String(readableType(identity.Type)),
		identity.Name,
		caser.String(strings.Join(readableRoles(identity.Roles), ", ")),
	}
}

func RowForProjectIAMIdentity(projectID string) func(resources.IAMIdentity) []string {
	return func(identity resources.IAMIdentity) []string {
		fmt.Println(identity)
		var roleStrings []string
		inherited := ""
		caser := cases.Title(language.English)
		var environmentRoles []string

		if len(identity.ProjectsRole) > 0 {
			for _, role := range identity.ProjectsRole {
				if role.ID == projectID {
					roleStrings = readableRoles(role.Roles)
				}
				if len(role.Environments) > 0 {
					for _, environment := range role.Environments {
						environmentRoles = append(environmentRoles, environment.ID+"="+caser.String(strings.Join(readableRoles(environment.Roles), ", ")))
					}
				}
			}
		}

		if len(roleStrings) == 0 {
			roleStrings = projectReadableRoles(identity.Roles)
			inherited = " (inherited)"
		}

		roles := caser.String(strings.Join(roleStrings, ", "))
		if len(roles) > 0 {
			roles += inherited
		}
		return []string{
			identity.ID,
			caser.String(readableType(identity.Type)),
			identity.Name,
			roles,
			strings.Join(environmentRoles, " "),
		}
	}
}

func RowForUserIdentity(identity resources.UserIdentity) []string {
	caser := cases.Title(language.English)
	groupNames := make([]string, 0, len(identity.Groups))
	for _, group := range identity.Groups {
		groupNames = append(groupNames, group.Name)
	}

	groups := "-"
	if len(groupNames) > 0 {
		groups = strings.Join(groupNames, ", ")
	}

	roles := "-"
	if len(identity.Roles) > 0 {
		roles = caser.String(strings.Join(readableRoles(identity.Roles), ", "))
	}

	lastLogin := "-"
	if !identity.LastLogin.IsZero() {
		lastLogin = util.HumanDuration(time.Since(identity.LastLogin))
	}

	return []string{
		identity.ID,
		identity.FullName,
		identity.Email,
		roles,
		groups,
		lastLogin,
	}
}

func RowForGroupIdentity(identity resources.GroupIdentity) []string {
	caser := cases.Title(language.English)
	memberNames := make([]string, 0, len(identity.Members))
	for _, member := range identity.Members {
		memberNames = append(memberNames, member.Name)
	}

	names := "-"
	if len(memberNames) > 0 {
		names = strings.Join(memberNames, ", ")
	}
	return []string{
		identity.ID,
		readableType(identity.Name),
		caser.String(readableRole(identity.Role)),
		names,
	}
}

func RowForServiceAccountIdentity(identity resources.ServiceAccountIdentity) []string {
	caser := cases.Title(language.English)
	roles := "-"
	if len(identity.Roles) > 0 {
		roles = caser.String(strings.Join(readableRoles(identity.Roles), ", "))
	}

	lastLogin := "-"
	if !identity.LastLogin.IsZero() {
		lastLogin = util.HumanDuration(time.Since(identity.LastLogin))
	}

	return []string{
		identity.ID,
		identity.Name,
		roles,
		lastLogin,
	}
}

func readableType(identityType string) string {
	switch identityType {
	case UsersEntityName:
		return "user"
	case GroupsEntityName:
		return "group"
	case ServiceAccountsEntityName:
		return "service account"
	default:
		return identityType
	}
}

func readableRoles(roles []string) []string {
	transformedRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		transformedRoles = append(transformedRoles, readableRole(role))
	}

	return transformedRoles
}

func projectReadableRoles(roles []string) []string {
	transformedRoles := make([]string, 0)
	for _, role := range roles {
		if role != "guest" {
			transformedRoles = append(transformedRoles, readableRole(role))
		}
	}

	return transformedRoles
}

func readableRole(role string) string {
	switch role {
	case "company-owner":
		return "company owner"
	case "project-admin":
		return "project administrator"
	default:
		return role
	}
}
