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

import "github.com/mia-platform/miactl/internal/resources"

type roleChanges struct {
	companyID       string
	projectID       string
	entityID        string
	entityType      string
	environmentName string
	environmentRole string
	projectRole     *string
}

func payloadForChanges(changes roleChanges) resources.EditIAMRole {
	projectRoles := resources.EditProjectRole{
		ProjectID: changes.projectID,
	}

	if changes.projectRole != nil {
		role := resources.IAMRole(*changes.projectRole)
		projectRoles.Role = &role
	}

	if len(changes.environmentName) > 0 {
		projectRoles.EnvironmentsRole = []resources.EditEnvironmentRole{
			{
				EnvironmentID: changes.environmentName,
				Role:          resources.IAMRole(changes.environmentRole),
			},
		}
	}

	return resources.EditIAMRole{
		ProjectsRole: []resources.EditProjectRole{projectRoles},
	}
}
