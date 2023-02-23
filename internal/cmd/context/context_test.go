package context

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewContextCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		cmd := NewContextCmd()
		require.NotNil(t, cmd)
	})
}
