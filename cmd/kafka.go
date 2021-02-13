package cmd

import (
	"github.com/spf13/cobra"
)

// TODO: deve essere dichiarativa come lo yaml di k8s

// NewGetCmd func creates a new command
func newKafkaCmd() *cobra.Command {

	var validKafkaArgs = []string{
		"topic",
	}

	kafkaCommand := &cobra.Command{
		Short:     "Manage Mia-Platform Kafka cluster",
		Long:      "",
		Use:       "kafka",
		ValidArgs: validKafkaArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "topic":
				cmd.MarkFlagRequired("topic")
			}
			return nil
		},
	}

	// add sub command to root command
	kafkaCommand.AddCommand(NewKafkaTopic())
	return kafkaCommand
}
