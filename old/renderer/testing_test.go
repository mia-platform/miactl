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
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanTableRows(t *testing.T) {
	buf := &bytes.Buffer{}

	table := NewTable(buf, []string{"h1", "h2", "h3"})
	table.AppendBulk([][]string{
		{"r1c1", "r1c2", "r1c3"},
		{"r2c1", "r2c2", "r2c3"},
		{"r3c1", "r3c2", "r3c3"},
	})
	table.Render()

	cleaned := CleanTableRows(buf.String())
	require.Equal(t, []string{
		"H1 | H2 | H3",
		"r1c1 | r1c2 | r1c3",
		"r2c1 | r2c2 | r2c3",
		"r3c1 | r3c2 | r3c3",
	}, cleaned)
}
