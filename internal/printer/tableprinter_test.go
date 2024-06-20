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

package printer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTablePrinter(t *testing.T) {
	t.Run("simple table print", func(t *testing.T) {
		str := &strings.Builder{}
		p := NewTablePrinter(TablePrinterOptions{}).
			SetWriter(str).
			Keys("k1", "k2").
			Record("d1", "d2")
		p.Print()

		expected := `  K1  K2  

  d1  d2  
`

		require.Equal(
			t,
			expected,
			str.String(),
		)
	})
}
