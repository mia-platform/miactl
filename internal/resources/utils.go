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
	IAMRoleCompanyOwner,
}

func (role IAMRole) String() string {
	return string(role)
}

func IsValidIAMRole(role IAMRole) bool {
	for _, validRole := range validServiceAccountRoles {
		if validRole == role {
			return true
		}
	}

	return false
}

func IAMRoleCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	allRoles := []string{
		IAMRoleGuest.String(),
		IAMRoleReporter.String(),
		IAMRoleDeveloper.String(),
		IAMRoleMaintainer.String(),
		IAMRoleProjectAdmin.String(),
		IAMRoleCompanyOwner.String(),
	}

	if len(toComplete) == 0 {
		return allRoles, cobra.ShellCompDirectiveDefault
	}

	var completableRole []string
	for _, role := range allRoles {
		if strings.HasPrefix(role, toComplete) {
			completableRole = append(completableRole, role)
		}
	}

	return completableRole, cobra.ShellCompDirectiveDefault
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
