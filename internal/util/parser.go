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

package util

import "github.com/mia-platform/miactl/internal/client"

func RowsForResources[T any](response *client.Response, rowParser func(T) []string) ([][]string, error) {
	resources := make([]T, 0)
	if err := response.ParseResponse(&resources); err != nil {
		return nil, err
	}

	rows := make([][]string, 0)
	for _, resource := range resources {
		rows = append(rows, rowParser(resource))
	}
	return rows, nil
}
