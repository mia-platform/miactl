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

package renderer

import (
	"fmt"
	"strings"
)

// CleanTableRows is a test function to be used, to returns an array
// of rows. It is an utility to assert easier the result of the renderer function
func CleanTableRows(s string) []string {
	rows := strings.Split(strings.TrimSpace(s), "\n")

	cleanRows := []string{}

	for _, row := range rows {
		cleanCells := ""
		for _, cell := range strings.Split(strings.TrimSpace(row), "\t") {
			cleanCells += fmt.Sprintf("%s | ", strings.TrimSpace(cell))
		}
		cleanRows = append(cleanRows, strings.TrimSuffix(cleanCells, " | "))
	}
	return cleanRows
}
