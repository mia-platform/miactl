package auth

import (
	"github.com/mia-platform/miactl/cmd/auth/login"

	"github.com/spf13/cobra"
)

// NewAuthCmd create a new auth command
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "login",
	}

	cmd.AddCommand(login.NewLoginCmd())

	return cmd
}
