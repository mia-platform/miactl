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
	ID string `json:"id" yaml:"id"`
}
type Languages string

// TODO: Constraint type on these values
const (
	En Languages = "en"
	It Languages = "it"
)

type IntlMessages map[Languages]string

type Visibility struct {
	ContextType Context `json:"contextType" yaml:"contextType"`
	ContextID   string  `json:"contextId" yaml:"contextId"`
}

type Category struct {
	ID        string       `json:"id" yaml:"id"`
	LabelIntl IntlMessages `json:"labelIntl,omitempty" yaml:"labelIntl,omitempty"`
}

type Menu struct {
	ID        string       `json:"id" yaml:"id"`
	LabelIntl IntlMessages `json:"labelIntl" yaml:"labelIntl"`
	Order     *Order       `json:"order,omitempty" yaml:"order,omitempty"`
}

type ExtensionInfo struct {
	ExtensionID        string          `json:"extensionId" yaml:"extensionId"`
	Name               string          `json:"name" yaml:"name"`
	Entry              string          `json:"entry" yaml:"entry"`
	Type               string          `json:"type" yaml:"type"`
	Destination        DestinationArea `json:"destination" yaml:"destination"`
	Description        string          `json:"description,omitempty" yaml:"description,omitempty"`
	IconName           string          `json:"iconName,omitempty" yaml:"iconName,omitempty"`
	ActivationContexts []Context       `json:"activationContexts" yaml:"activationContexts"`
	Permissions        []string        `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	Visibilities       []Visibility    `json:"visibilities,omitempty" yaml:"visibilities,omitempty"`
	Category           Category        `json:"category,omitempty" yaml:"category,omitempty"`
	Menu               Menu            `json:"menu" yaml:"menu"`
}
