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

	"github.com/stretchr/testify/assert"

	"github.com/mia-platform/miactl/internal/resources"
)

func TestPayloadForChanges(t *testing.T) {
	projectID := "project-id"
	environmentID := "environment-id"
	projectRole := string(resources.IAMRoleProjectAdmin)
	iamProjectRole := resources.IAMRole(projectRole)
	environmenRole := string(resources.IAMRoleGuest)
	emptyRole := ""
	emptyIAMRole := resources.IAMRole(emptyRole)
	testCases := map[string]struct {
		expectedPayload resources.EditIAMRole
		changes         roleChanges
	}{
		"change only project role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      &iamProjectRole,
					},
				},
			},
			changes: roleChanges{
				projectID:   projectID,
				projectRole: &projectRole,
			},
		},
		"change only environment role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          resources.IAMRole(environmenRole),
							},
						},
					},
				},
			},
			changes: roleChanges{
				projectID:       projectID,
				environmentName: environmentID,
				environmentRole: environmenRole,
			},
		},
		"both roles": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      &iamProjectRole,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          resources.IAMRole(environmenRole),
							},
						},
					},
				},
			},
			changes: roleChanges{
				projectID:       projectID,
				projectRole:     &projectRole,
				environmentName: environmentID,
				environmentRole: environmenRole,
			},
		},
		"remove project Role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      &emptyIAMRole,
					},
				},
			},
			changes: roleChanges{
				projectID:   projectID,
				projectRole: &emptyRole,
			},
		},
		"remove environment Role": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          emptyIAMRole,
							},
						},
					},
				},
			},
			changes: roleChanges{
				projectID:       projectID,
				environmentName: environmentID,
				environmentRole: emptyRole,
			},
		},
		"remove both roles": {
			expectedPayload: resources.EditIAMRole{
				ProjectsRole: []resources.EditProjectRole{
					{
						ProjectID: projectID,
						Role:      &emptyIAMRole,
						EnvironmentsRole: []resources.EditEnvironmentRole{
							{
								EnvironmentID: environmentID,
								Role:          emptyIAMRole,
							},
						},
					},
				},
			},
			changes: roleChanges{
				projectID:       projectID,
				projectRole:     &emptyRole,
				environmentName: environmentID,
				environmentRole: emptyRole,
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			payload := payloadForChanges(testCase.changes)
			assert.Equal(t, testCase.expectedPayload, payload)
		})
	}
}
