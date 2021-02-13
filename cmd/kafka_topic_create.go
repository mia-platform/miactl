package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"
)

var Partitions int
var Gracefully bool

// NewKafkaTopicCreate creates a topic for Kafka
func NewKafkaTopicCreate() *cobra.Command {

	createTopicCmd := &cobra.Command{
		Short: "Create a Kafka Topic",
		Long:  "Use this command to create a Kafka Topic on a Mia-Platform cluster",
		Use:   "create <topic name>",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := GetFactoryFromContext(cmd.Context(), opts)
			if err != nil {
				return err
			}
			createTopic(f, args)
			return nil
		},
	}

	createTopicCmd.Flags().IntVarP(&Partitions, "partitions", "", 1, "Number of topic partitions.")
	createTopicCmd.Flags().BoolVarP(&Gracefully, "if-not-exists", "", false, "Exit gracefully if topic already exists.")

	return createTopicCmd
}

func createTopic(f *Factory, args []string) {

	broker := args[1]
	topic := args[2]
	numParts, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Printf("Invalid partition count: %s: %v\n", os.Args[3], err)
		return
	}
	replicationFactor, err := strconv.Atoi(args[4])
	if err != nil {
		fmt.Printf("Invalid replication factor: %s: %v\n", os.Args[4], err)
		return
	}

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
