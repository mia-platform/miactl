package console

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConsoleCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		cmd := NewConsoleCmd()
		require.NotNil(t, cmd)
	})
}
