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

package resources

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

var validServiceAccountRoles = []IAMRole{
	IAMRoleGuest,
	IAMRoleReporter,
	IAMRoleDeveloper,
	IAMRoleMaintainer,
	IAMRoleProjectAdmin,
}

func (role IAMRole) String() string {
	return string(role)
}

func IsValidIAMRole(role IAMRole, project bool) bool {
	roles := validServiceAccountRoles
	if !project {
		roles = append(roles, IAMRoleCompanyOwner)
	}
	for _, validRole := range roles {
		if validRole == role {
			return true
		}
	}

	return false
}

func IsValidEnvironmentRole(role IAMRole) bool {
	roles := []IAMRole{
		IAMRoleReporter,
		IAMRoleMaintainer,
	}

	for _, validRole := range roles {
		if validRole == role {
			return true
		}
	}

	return false
}

func IAMRoleCompletion(project bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	allRoles := []string{
		string(IAMRoleGuest),
		string(IAMRoleReporter),
		string(IAMRoleDeveloper),
		string(IAMRoleMaintainer),
		string(IAMRoleProjectAdmin),
	}

	if !project {
		allRoles = append(allRoles, string(IAMRoleCompanyOwner))
	}

	return completionForRoles(allRoles)
}

func IAMEnvironmentRoleCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completionForRoles([]string{string(IAMRoleReporter), string(IAMRoleMaintainer)})
}

func completionForRoles(roles []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(toComplete) == 0 {
			return roles, cobra.ShellCompDirectiveDefault
		}

		var completableRole []string
		for _, role := range roles {
			if strings.HasPrefix(role, toComplete) {
				completableRole = append(completableRole, role)
			}
		}

		return completableRole, cobra.ShellCompDirectiveDefault
	}
}

func EncodeResourceToJSON(obj interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

type JWTServiceAccountJSON struct {
	Type           string `json:"type"`
	KeyID          string `json:"key-id"`           //nolint: tagliatelle
	PrivateKeyData string `json:"private-key-data"` //nolint: tagliatelle
	ClientID       string `json:"client-id"`        //nolint: tagliatelle
}
