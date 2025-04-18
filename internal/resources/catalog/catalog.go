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

package catalog

import (
	"errors"

	"github.com/mia-platform/miactl/internal/encoding"
)

var (
	ErrItemNotFound          = errors.New("item not found")
	ErrVersionNameNotAString = errors.New(`the field "version.name" must be a string`)
	ErrMissingCompanyID      = errors.New("missing company id, please set one with the flag company-id or in the context")
)

// Item is a Catalog item
// we use a map[string]interface{} to represent the item
// this allows to avoid changes in the code in case of a change in the resource structure
type Item map[string]interface{}

type ApplyResponse struct {
	Done  bool                `json:"done"`
	Items []ApplyResponseItem `json:"items"`
}

type ApplyResponseItem struct {
	ID     string `json:"_id,omitempty"` //nolint: tagliatelle
	ItemID string `json:"itemId,omitempty"`

	Done     bool `json:"done"`
	Inserted bool `json:"inserted"`
	Updated  bool `json:"updated"`

	Errors []ApplyResponseItemError `json:"errors"`
}

type ApplyResponseItemError struct {
	Message string `json:"message"`
}

type ApplyRequest struct {
	Resources []*Item `json:"resources"`
}

type UploadImageResponse struct {
	ID       string `json:"_id"` //nolint: tagliatelle
	Name     string `json:"name"`
	File     string `json:"file"`
	Size     int64  `json:"size"`
	Location string `json:"location"`
}

type Release struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (i *Item) Marshal(encodingFormat string) ([]byte, error) {
	return encoding.MarshalData(i, encodingFormat, encoding.MarshalOptions{Indent: true})
}

func (i *Item) Del(key string) {
	delete(*i, key)
}

func (i *Item) Set(key string, val interface{}) {
	(*i)[key] = val
}

func (i *Item) Get(key string) interface{} {
	return (*i)[key]
}

func (i *Item) GetVersionName() (versionName string, err error) {
	version, ok := i.Get("version").(Item)
	if ok && version != nil {
		versionName, ok = version["name"].(string)
		if !ok {
			return "", ErrVersionNameNotAString
		}
	}
	return
}
