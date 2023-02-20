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
