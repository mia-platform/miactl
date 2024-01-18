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

	"github.com/mia-platform/miactl/internal/client"
)

func ListAllIAMEntities(ctx context.Context, client *client.APIClient, companyID string, projectIds []string, entityTypes map[string]bool) (*client.Response, error) {
	request := client.
		Get().
		APIPath(fmt.Sprintf(entititesPathTemplate, companyID))

	if len(projectIds) > 0 {
		request.SetParam(projectIdsKey, projectIds...)
	}

	for entityName, enabled := range entityTypes {
		if !enabled {
			continue
		}
		request.SetParam("identityType", entityName)
	}

	return request.Do(ctx)
}

func ListSpecificEntities(ctx context.Context, client *client.APIClient, companyID string, entityType string) (*client.Response, error) {
	var apiPathTemplate string

	switch entityType {
	case UsersEntityName:
		apiPathTemplate = usersPathTemplate
	case GroupsEntityName:
		apiPathTemplate = groupsPathTemplate
	case ServiceAccountsEntityName:
		apiPathTemplate = serviceAccountsPathTemplate
	default:
		return nil, fmt.Errorf("unknown IAM entity")
	}

	response, err := client.
		Get().
		APIPath(fmt.Sprintf(apiPathTemplate, companyID)).
		Do(ctx)
	return response, err
}
