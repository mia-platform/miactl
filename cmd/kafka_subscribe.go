package cmd

import (
	"github.com/spf13/cobra"
)

// NewKafkaSubscribeTopic subscribe to a Kafka topic and shows the messages on it
func NewKafkaSubscribeTopic() *cobra.Command {

	var validKafkaArgs = []string{
		"topic",
	}

	return &cobra.Command{
		Short:     "Subscribe to Kafka topic",
		Long:      "",
		Use:       "subscribe",
		ValidArgs: validKafkaArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "subscribe":
				cmd.MarkFlagRequired("subscribe")
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
			case "subscribe":
				subscribeTopic(f, args)
			}
			return nil
		},
	}
}

func subscribeTopic(f *Factory, args []string) {
	return
}
