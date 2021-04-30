package console

import (
	"github.com/mia-platform/miactl/cmd/console/login"

	"github.com/spf13/cobra"
)

// NewConsoleCmd create a new Console command
func NewConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "console <command>",
	}

	cmd.AddCommand(login.NewLoginCmd())

	return cmd
}
