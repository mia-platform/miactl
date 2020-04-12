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
		[]string{"r1c1", "r1c2", "r1c3"},
		[]string{"r2c1", "r2c2", "r2c3"},
		[]string{"r3c1", "r3c2", "r3c3"},
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
