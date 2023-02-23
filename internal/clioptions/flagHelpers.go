package clioptions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetFlagString(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		fmt.Printf("%v", err)
	}
	return s
}
