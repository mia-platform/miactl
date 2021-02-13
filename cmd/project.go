package cmd

import (
	"github.com/spf13/cobra"
)

// NewGetCmd func creates a new command
func newProjectCmd() *cobra.Command {

	var validKafkaArgs = []string{
		"get", "gets",
	}

	projectCommand := &cobra.Command{
		Short:     "Manage Mia-Platform Projects",
		Long:      "",
		Use:       "project",
		ValidArgs: validKafkaArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
	}

	projectCommand.PersistentFlags().StringVarP(&projectID, "project", "p", "", "specify desired project ID")

	// add sub command to root command
	projectCommand.AddCommand(newGetCmd())

	return projectCommand
}
