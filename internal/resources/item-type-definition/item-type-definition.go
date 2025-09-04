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

package itd

import (
	"errors"

	"github.com/mia-platform/miactl/internal/encoding"
)

var (
	ErrUnsupportedCompanyVersion = errors.New("you need Mia-Platform Console v14.1.0 or later to use this command")
	ErrMissingCompanyID          = errors.New("missing company id, please set one with the flag company-id or in the context")
	ErrItemNotFound              = errors.New("item type definition not found")
)

// Item is a Marketplace item
// we use a map[string]interface{} to represent the item
// this allows to avoid changes in the code in case of a change in the resource structure
type GenericItemTypeDefinition map[string]interface{}

type ItemTypeDefinitionMetadataVisibility struct {
	Scope string   `json:"scope"`
	Ids   []string `json:"ids"`
}

type ItemTypeDefinitionMetadataNamespace struct {
	Scope string `json:"scope"`
	Id    string `json:"id"`
}

type ItemTypeDefinitionMetadataPublisher struct {
	Name string `json:"name"`
}

type ItemTypeDefinitionMetadata struct {
	Name        string                               `json:"name"`
	DisplayName string                               `json:"displayName"`
	Namespace   ItemTypeDefinitionMetadataNamespace  `json:"namespace"`
	Visibility  ItemTypeDefinitionMetadataVisibility `json:"visibility"`
	Publisher   ItemTypeDefinitionMetadataPublisher  `json:"publisher"`
}

type ItemTypeDefinitionSpec struct {
	IsVersioningSupported bool `json:"isVersioningSupported"`
}

type ItemTypeDefinition struct {
	Metadata ItemTypeDefinitionMetadata `json:"metadata"`
	Spec     ItemTypeDefinitionSpec     `json:"spec"`
}

func (i *GenericItemTypeDefinition) Marshal(encodingFormat string) ([]byte, error) {
	return encoding.MarshalData(i, encodingFormat, encoding.MarshalOptions{Indent: true})
}

func (i *GenericItemTypeDefinition) Del(key string) {
	delete(*i, key)
}

func (i *GenericItemTypeDefinition) Set(key string, val interface{}) {
	(*i)[key] = val
}

func (i *GenericItemTypeDefinition) Get(key string) interface{} {
	return (*i)[key]
}
