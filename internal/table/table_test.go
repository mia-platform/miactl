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

package table

import (
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestTable(t *testing.T) {
	t.Run("simple table", func(t *testing.T) {
		str := &strings.Builder{}
		New(str).Header("C1", "C2").Row("D1", "D2").Print()
		snaps.MatchSnapshot(t, str.String())
	})
}
