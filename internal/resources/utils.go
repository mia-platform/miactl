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
)

var validServiceAccountRoles = []ServiceAccountRole{
	ServiceAccountRoleGuest,
	ServiceAccountRoleReporter,
	ServiceAccountRoleDeveloper,
	ServiceAccountRoleMaintainer,
	ServiceAccountRoleProjectAdmin,
	ServiceAccountRoleCompanyOwner,
}

func (role ServiceAccountRole) String() string {
	return string(role)
}

func IsValidServiceAccountRole(role ServiceAccountRole) bool {
	for _, validRole := range validServiceAccountRoles {
		if validRole == role {
			return true
		}
	}

	return false
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
