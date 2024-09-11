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

type Order float64

type Context string

type DestinationArea struct {
	ID   string `yaml:"id" json:"id" `
	Path string `yaml:"path" json:"path"`
}
type Languages string

// TODO: Constraint type on these values
const (
	En Languages = "en"
	It Languages = "it"
)

type IntlMessages map[Languages]string

type Visibility struct {
	ContextType Context `yaml:"contextType" json:"contextType"`
	ContextID   string  `yaml:"contextId" json:"contextId"`
}

type Category struct {
	ID        string       `yaml:"id" json:"id" `
	LabelIntl IntlMessages `yaml:"labelIntl,omitempty" json:"labelIntl,omitempty"`
	Order     *Order       `yaml:"order,omitempty" json:"order,omitempty"`
}

type Menu struct {
	ID        string       `yaml:"id" json:"id"`
	LabelIntl IntlMessages `yaml:"labelIntl" json:"labelIntl"`
	Order     *Order       `yaml:"order,omitempty" json:"order,omitempty"`
}

type ExtensionInfo struct {
	ExtensionID        string          `yaml:"extensionId" json:"extensionId"`
	Name               string          `yaml:"name" json:"name"`
	Entry              string          `yaml:"entry" json:"entry"`
	Type               string          `yaml:"type" json:"type"`
	Destination        DestinationArea `yaml:"destination" json:"destination"`
	Description        string          `yaml:"description,omitempty" json:"description,omitempty"`
	IconName           string          `yaml:"iconName,omitempty" json:"iconName,omitempty"`
	ActivationContexts []Context       `yaml:"activationContexts" json:"activationContexts"`
	Permissions        []string        `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Visibilities       []Visibility    `yaml:"visibilities,omitempty" json:"visibilities,omitempty"`
	Category           *Category       `yaml:"category,omitempty" json:"category,omitempty"`
	Menu               *Menu           `yaml:"menu" json:"menu,omitempty"`
	//nolint:tagliatelle
	RoleIDs []string `yaml:"roleIds,omitempty" json:"roleIds,omitempty"`
}
