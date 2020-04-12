package renderer

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// NewTable create a new table without border
func NewTable(writer io.Writer, headersTitle []string) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)

	table.SetHeader(headersTitle)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	return table
}
