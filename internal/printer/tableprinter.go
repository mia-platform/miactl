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
	"io"

	"github.com/olekukonko/tablewriter"
)

type TablePrinterOptions struct {
	WrapLinesDisabled bool
}

type TablePrinter struct {
	w       io.Writer
	tw      *tablewriter.Table
	options TablePrinterOptions
}

func (t *TablePrinter) commonTableSetup(tw *tablewriter.Table) {
	tw.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetCenterSeparator("")
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetAutoWrapText(!t.options.WrapLinesDisabled)
}

func NewTablePrinter(options TablePrinterOptions) *TablePrinter {
	return &TablePrinter{options: options}
}

func (t *TablePrinter) SetWriter(w io.Writer) IPrinter {
	table := tablewriter.NewWriter(w)
	t.commonTableSetup(table)

	t.w = w
	t.tw = table

	return t
}

func (t *TablePrinter) Keys(keys ...string) IPrinter {
	t.tw.SetHeader(keys)
	return t
}

func (t *TablePrinter) Record(recordValues ...string) IPrinter {
	t.tw.Append(recordValues)
	return t
}

func (t *TablePrinter) BulkRecords(records ...[]string) IPrinter {
	t.tw.AppendBulk(records)
	return t
}

func (t *TablePrinter) Print() {
	t.tw.Render()
}
