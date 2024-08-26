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
	ExtensionID        string          `yaml:"extensionId,omitempty" json:"extensionId,omitempty"`
	Name               string          `yaml:"name" json:"name"`
	Entry              string          `yaml:"entry" json:"entry"`
	Type               string          `yaml:"type" json:"type"`
	Destination        DestinationArea `yaml:"destination" json:"destination"`
	Description        string          `yaml:"description,omitempty" json:"description,omitempty"`
	IconName           string          `yaml:"iconName,omitempty" json:"iconName,omitempty"`
	ActivationContexts []Context       `yaml:"activationContexts" json:"activationContexts"`
	Permissions        []string        `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Roles              []string        `yaml:"roleIds,omitempty" json:"roleIds,omitempty"`
	Category           *Category       `yaml:"category,omitempty" json:"category,omitempty"`
	Menu               Menu            `yaml:"menu" json:"menu"`
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
