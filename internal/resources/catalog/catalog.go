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

	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

var (
	ErrUnsupportedCompanyVersion = errors.New("you need Mia-Platform Console v14.0.0 or later to use this command")
)

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
	Resources []*marketplace.Item `json:"resources"`
}
