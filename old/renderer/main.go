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
	"io"

	"github.com/olekukonko/tablewriter"
)

// IRenderer is the exported renderer interface
type IRenderer interface {
	Error(err error) IError
	Table(headersString []string) *tablewriter.Table
}

// Renderer implementation of IRenderer interface
type Renderer struct {
	writer io.Writer
}

// Error method create a new error writer
func (r *Renderer) Error(err error) IError {
	return NewError(r.writer, err)
}

// Table method create a new table writer
func (r *Renderer) Table(headersString []string) *tablewriter.Table {
	return NewTable(r.writer, headersString)
}

// New create the renderer implementation
func New(writer io.Writer) IRenderer {
	return &Renderer{
		writer: writer,
	}
}
