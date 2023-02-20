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
