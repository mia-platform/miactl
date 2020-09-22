package auth

import (
	"github.com/mia-platform/miactl/cmd/auth/login"
	"github.com/mia-platform/miactl/sdk"

	"github.com/spf13/cobra"
)

func NewAuthCmd(opts sdk.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "login",
	}

	cmd.AddCommand(login.NewLoginCmd(opts))

	return cmd
}
