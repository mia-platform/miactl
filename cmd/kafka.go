package cmd

import (
	"github.com/spf13/cobra"
)

// NewGetCmd func creates a new command
func newKafkaCmd() *cobra.Command {

	var validKafkaArgs = []string{
		"create",
		"subscribe",
		"produce",
	}

	return &cobra.Command{
		Short:     "Manage Mia-Platform Kafka cluster",
		Long:      "",
		Use:       "kafka",
		ValidArgs: validKafkaArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "create":
			case "subscribe":
			case "produce":
				cmd.MarkFlagRequired("topic")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := GetFactoryFromContext(cmd.Context(), opts)
			if err != nil {
				return err
			}

			resource := args[0]

			switch resource {
			case "create":
				createTopic(f)
			case "subscribe":
				subscribeTopic(f)
			case "produce":
				produceMessage(f)
			}
			return nil
		},
	}
}

func createTopic(f *Factory) {
	return
}

func subscribeTopic(f *Factory) {
	return
}

func produceMessage(f *Factory) {
	return
}
