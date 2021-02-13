package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	t.Run("without correct args", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "not-correct-arg")
		expectedErrMessage := `invalid argument "not-correct-arg" for "miactl completion"`
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without args", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion")
		expectedErrMessage := `accepts 1 arg(s), received 0`
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("with fish arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "fish")
		require.Nil(t, err)
		require.Contains(t, out, "# fish completion for miactl")
	})

	t.Run("with bash arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "bash")
		require.Nil(t, err)
		require.Contains(t, out, "# bash completion for miactl")
	})

	t.Run("with zsh arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "zsh")
		require.Nil(t, err)
		require.Contains(t, out, "#compdef _miactl miactl")
	})
}
