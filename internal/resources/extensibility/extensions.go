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

package extensibility

type Extension struct {
	ExtensionID   string            `yaml:"extensionId,omitempty" json:"extensionId,omitempty"`
	ExtensionType string            `yaml:"extensionType,omitempty" json:"extensionType,omitempty"`
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	Entry         string            `yaml:"entry" json:"entry"`
	Contexts      []string          `yaml:"contexts" json:"contexts"`
	Permissions   []string          `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Routes        []*ExtensionRoute `yaml:"routes,omitempty" json:"routes,omitempty"`
}

type Icon struct {
	Name string `json:"name"`
}

type ExtensionRoute struct {
	ID              string            `yaml:"id" json:"id"`
	LocationID      string            `yaml:"locationId" json:"locationId"`
	ParentID        string            `yaml:"parentId,omitempty" json:"parentId,omitempty"`
	DestinationPath string            `yaml:"destinationPath,omitempty" json:"destinationPath,omitempty"`
	LabelIntl       map[string]string `yaml:"labelIntl,omitempty" json:"labelIntl,omitempty"`
	RenderType      string            `yaml:"renderType,omitempty" json:"renderType,omitempty"`
	Order           *float64          `yaml:"order,omitempty" json:"order,omitempty"`
	Icon            *Icon             `yaml:"icon,omitempty" json:"icon,omitempty"`
}
