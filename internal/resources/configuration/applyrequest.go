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

package configuration

const (
	DefaultTitle = "[miactl] Applied project configuration"
)

type ApplyRequest struct {
	*Configuration

	Title        string `json:"title" yaml:"title"`
	PreviousSave string `json:"previousSave,omitempty" yaml:"previousSave,omitempty"`
}

func BuildApplyRequest(config *Configuration) *ApplyRequest {
	req := &ApplyRequest{
		Configuration: config,
	}

	return req.
		withPreviousSnapshotID(config.Config["commitId"].(string)).
		WithTitle(DefaultTitle)
}

func (r *ApplyRequest) WithTitle(title string) *ApplyRequest {
	r.Title = title
	return r
}

func (r *ApplyRequest) withPreviousSnapshotID(previousSnapshotID string) *ApplyRequest {
	r.PreviousSave = previousSnapshotID
	return r
}
