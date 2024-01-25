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
	"github.com/mia-platform/miactl/internal/resources"
)

func EditIAMResourceRole(ctx context.Context, client *client.APIClient, companyID, resourceID, entityType string, change resources.EditIAMRole) (*client.Response, error) {
	apiPath := ""
	switch entityType {
	case UsersEntityName:
		apiPath = usersPathTemplate
	case GroupsEntityName:
		apiPath = groupsPathTemplate
	case ServiceAccountsEntityName:
		apiPath = serviceAccountsPathTemplate
	default:
		return nil, fmt.Errorf("unknown IAM entity: %q", entityType)
	}

	apiPath += "/%s"

	payload, err := resources.EncodeResourceToJSON(change)
	if err != nil {
		return nil, err
	}

	return client.Patch().
		APIPath(fmt.Sprintf(apiPath, companyID, resourceID)).
		Body(payload).
		Do(ctx)
}
