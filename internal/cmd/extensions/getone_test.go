package extensions

import (
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"

	"github.com/stretchr/testify/require"
)

func TestGetOneCmdCommandBuilder(t *testing.T) {
	opts := clioptions.NewCLIOptions()
	cmd := GetOneCmd(opts)
	require.NotNil(t, cmd)
}
