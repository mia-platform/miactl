package context

import (
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func NewContextCmd() *cobra.Command {
	options := clioptions.NewRootOptions()
	cmd := &cobra.Command{
		Use:   "context",
		Short: "perform operations on cluster contexts",
	}

	cmd.AddCommand(NewSetContextCmd(options))
	cmd.AddCommand(NewUseContextCmd(options))

	options.AddFlags(cmd)

	return cmd
}
