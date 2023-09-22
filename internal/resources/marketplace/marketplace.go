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

package marketplace

import "github.com/mia-platform/miactl/internal/encoding"

// Item is a Marketplace item
// we use a map[string]interface{} to represent the item
// this allows to avoid changes in the code in case of a change in the resource structure
type Item map[string]interface{}

type ApplyResponse struct {
	Done  bool                `json:"done"`
	Items []ApplyResponseItem `json:"items"`
}

type ApplyResponseItem struct {
	ItemID string `json:"itemId,omitempty"`
	Name   string `json:"name,omitempty"`

	Done     bool `json:"done"`
	Inserted bool `json:"inserted"`
	Updated  bool `json:"updated"`

	ValidationErrors []ApplyResponseItemValidationError `json:"validationErrors"`
}

type ApplyResponseItemValidationError struct {
	Message string `json:"message"`
}

type ApplyRequest struct {
	Resources []Item `json:"resources"`
}

func (i *Item) MarshalItem(encodingFormat string) ([]byte, error) {
	return encoding.MarshalData(i, encodingFormat, encoding.MarshalOptions{Indent: true})
}
