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

/**
  "name": "Deployer helper",
  "entry": "https://example.com/",
  "contexts": ["project"],
  "routes": [{
    "id": "react-app",
    "parentId": "workloads",
    "locationId": "runtime",
    "labelIntl": {
      "en": "SomeLabel"
      "it": "SomeLabelInItalian"
    },
    "destinationPath": "/",
    "order": 200,
    "icon": {
      "name": "PiHardDrives"
    }
  }]
*/

type Extension struct {
	ExtensionID string            `json:"extensionId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Entry       string            `json:"entry"`
	Contexts    string            `json:"contexts"`
	Permissions []string          `json:"permissions,omitempty"`
	Routes      []*ExtensionRoute `json:"routes,omitempty"`
}

type Languages string

const (
	En Languages = "en"
	It Languages = "it"
)

type IntlMessages map[Languages]string

type Icon struct {
	Name string `json:"name"`
}

type ExtensionRoute struct {
	ID                  string       `json:"id"`
	ParentID            string       `json:"parentId,omitempty"`
	LocationID          string       `json:"locationId" jsonschema:"enum=tenant,enum=project,enum=runtime"`
	DestinationPath     string       `json:"destinationPath,omitempty"`
	LabelIntl           IntlMessages `json:"labelIntl"`
	MatchExactMountPath bool         `json:"matchExactMountPath,omitempty"`
	RenderType          string       `json:"renderType,omitempty" jsonschema:"enum=category"`
	Order               *float64     `json:"order,omitempty"`
	Icon                *Icon        `json:"icon,omitempty"`
}
