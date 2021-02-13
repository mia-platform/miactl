package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"
)

// NewKafkaTopicCreate creates a topic for Kafka
func NewKafkaTopicCreate() *cobra.Command {

	createTopicCmd := &cobra.Command{
		Short: "Create a Kafka Topic",
		Long: `Use this command to create a Kafka Topic on a Mia-Platform cluster.
Example:
  miactl kafka topic create --broker http://localhost:8999 --partitions 1  my-new-topic;`,
		Use:  "create <topic name>",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := GetFactoryFromContext(cmd.Context(), opts)
			if err != nil {
				return err
			}
			broker, _ := cmd.Flags().GetString("broker")
			topic := args[0]
			numParts, _ := cmd.Flags().GetInt("partitions")
			replicationFactor, _ := cmd.Flags().GetInt("replication")
			createTopic(f, broker, topic, numParts, replicationFactor)
			return nil
		},
	}

	createTopicCmd.Flags().IntP("partitions", "", 1, "number of topic partitions")
	createTopicCmd.Flags().IntP("replication", "", 1, "replication factor")
	createTopicCmd.Flags().BoolP("if-not-exists", "", false, "exit gracefully if topic already exists")

	return createTopicCmd
}

func createTopic(f *Factory, broker string, topic string, numParts int, replicationFactor int) {

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		return
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic("ParseDuration(60s)")
	}
	results, err := a.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numParts,
			ReplicationFactor: replicationFactor}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to create topic: %v\n", err)
		os.Exit(1)
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
}
