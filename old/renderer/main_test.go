package renderer

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create a IRenderer implementation", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := New(buf)
		require.Implements(t, (*IRenderer)(nil), r)
	})

	t.Run("Error method returns new error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := New(buf)
		err := fmt.Errorf("my error")
		expectedErr := NewError(buf, err)
		require.Equal(t, expectedErr, r.Error(err))
	})

	t.Run("Table method returns new table", func(t *testing.T) {
		headers := []string{"h1", "h2", "h3"}
		buf := &bytes.Buffer{}

		r := New(buf)
		r.Table(headers).Render()

		var expected bytes.Buffer
		NewTable(&expected, headers).Render()

		require.Equal(t, expected.String(), buf.String())
	})

	t.Run("render table with correct headers", func(t *testing.T) {
		var b bytes.Buffer
		headers := []string{"h1", "h2", "h3"}
		table := NewTable(&b, headers)
		table.Render()

		expectedStrings := "H1	H2	H3 \n"
		require.Equal(t, expectedStrings, b.String())
	})
}
