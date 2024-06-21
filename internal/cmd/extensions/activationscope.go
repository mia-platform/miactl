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

package extensions

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
)

const CompanyContext = "company"
const ProjectContext = "project"

type ActivationScope struct {
	ContextID   string `json:"contextId"`
	ContextType string `json:"contextType"`
}

func (a ActivationScope) String() string {
	return fmt.Sprintf("%s: %s", a.ContextType, a.ContextID)
}

func NewActivationScope(c *client.Config) ActivationScope {
	projectID := c.ProjectID
	if projectID != "" {
		return ActivationScope{
			ContextID:   c.ProjectID,
			ContextType: ProjectContext,
		}
	}
	return ActivationScope{
		ContextID:   c.CompanyID,
		ContextType: CompanyContext,
	}
}
