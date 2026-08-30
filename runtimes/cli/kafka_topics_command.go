package main

import (
	"fmt"

	"github.com/spf13/cobra"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	skafkatopics "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka/topics"
)

func init() {
	rootCommand.AddCommand(newKafkaCommand())
}

func newKafkaCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Kafka development resources.",
	}
	command.AddCommand(newEnsureKafkaTopicsCommand())
	return command
}

func newEnsureKafkaTopicsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "topics ensure",
		Short: "Create all versioned Notegic Kafka topics and their dead-letter topics.",
		RunE: func(command *cobra.Command, _ []string) error {
			connectionConfig, err := skafka.LoadConnectionConfig()
			if err != nil {
				return err
			}

			provisioner, err := skafka.NewTopicProvisioner(skafka.ClientConfig{
				ConnectionConfig: connectionConfig,
				ClientId:         "notegic-kafka-topic-bootstrap",
			})
			if err != nil {
				return err
			}
			defer provisioner.Close()

			if err := provisioner.EnsureTopics(command.Context(), skafkatopics.All()); err != nil {
				return fmt.Errorf("ensure Notegic Kafka topics: %w", err)
			}

			return nil
		},
	}
}
