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
