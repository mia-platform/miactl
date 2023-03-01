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

func TestNewTable(t *testing.T) {
	t.Run("render table with correct headers", func(t *testing.T) {
		var b bytes.Buffer
		headers := []string{"h1", "h2", "h3"}
		table := NewTable(&b, headers)
		table.Render()

		expectedStrings := "H1	H2	H3 \n"
		require.Equal(t, expectedStrings, b.String())
	})

	t.Run("render table with correct headers and data", func(t *testing.T) {
		var b bytes.Buffer
		headers := []string{"h1", "h2", "h3"}
		table := NewTable(&b, headers)
		table.Append([]string{"v1", "v2-data-long", "v3"})
		table.Render()

		expectedStrings := "H1	H2          	H3 \nv1	v2-data-long	v3	\n"
		require.Equal(t, expectedStrings, b.String())
	})
}
