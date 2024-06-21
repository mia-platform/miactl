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

package extensions

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type Table struct {
	tw *tablewriter.Table
	w  io.Writer
}

func commonTableSetup(tw *tablewriter.Table) {
	tw.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetCenterSeparator("")
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetAutoWrapText(true)
}

func NewTable(writer io.Writer) *Table {
	table := tablewriter.NewWriter(writer)
	commonTableSetup(table)

	return &Table{tw: table, w: writer}
}

func (t *Table) Header(header ...string) *Table {
	t.tw.SetHeader(header)
	return t
}

func (t *Table) Row(data ...string) *Table {
	t.tw.Append(data)
	return t
}

func (t *Table) Print() {
	t.tw.Render()
}
