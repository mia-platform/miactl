package console

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConsoleCommand(t *testing.T) {
	viper.SetConfigFile("/tmp/.miaplatformctl")

	cmd := NewConsoleCmd()

	require.Equal(t, "console <command>", cmd.Use)
}
