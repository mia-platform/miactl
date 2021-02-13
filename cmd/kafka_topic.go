package cmd

import (
	"github.com/spf13/cobra"
)



// NewKafkaTopic subscribe to a Kafka topic and shows the messages on it
func NewKafkaTopic() *cobra.Command {

	var validKafkaArgs = []string{
		"topic",
	}

	topicCommand := &cobra.Command{
		Short:     "Manage Kafka topic",
		Long:      "",
		Use:       "topic",
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

	topicCommand.AddCommand(NewKafkaTopicCreate())
	topicCommand.AddCommand(NewKafkaTopicSubscribe())
	topicCommand.AddCommand(NewKafkaTopicProduceMessage())

	return topicCommand
}
