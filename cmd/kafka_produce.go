package cmd

import (
	"github.com/spf13/cobra"
)

// NewKafkaProduceMessage produces a message to a Kafka topic
func NewKafkaProduceMessage() *cobra.Command {

	var validKafkaArgs = []string{
		"produce",
	}

	return &cobra.Command{
		Short: "Send a message to Kafka	",
		Long:      "",
		Use:       "produce",
		ValidArgs: validKafkaArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "produce":
				cmd.MarkFlagRequired("produce")
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
			case "produce":
				produceMessage(f, args)
			}
			return nil
		},
	}
}

func produceMessage(f *Factory, args []string) {
	return
}
