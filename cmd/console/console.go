package console

import (
	"github.com/mia-platform/miactl/cmd/console/deploy"
	"github.com/spf13/cobra"
)

func NewConsoleCmd() *cobra.Command {
	// Note: console should act as a resource that receives commands to be executed
	cmd := &cobra.Command{
		Use:   "console",
		Short: "select console resource",
	}

	cmd.AddCommand(deploy.NewDeployCmd())

	return cmd
}
