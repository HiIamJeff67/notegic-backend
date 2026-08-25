package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type KafkaConsumerConfig struct {
	Connection          skafka.ConnectionConfig
	ConsumerGroup       string
	MaximumAttempts     int
	InitialRetryBackoff time.Duration
	MaximumRetryBackoff time.Duration
	MaximumPollRecords  int
}

func (config KafkaConsumerConfig) ConsumerConfig() skafka.ConsumerConfig {
	return skafka.ConsumerConfig{
		ClientConfig: skafka.ClientConfig{
			ConnectionConfig: config.Connection,
			ClientId:         "notegic-notification-consumer",
		},
		ConsumerGroup:       config.ConsumerGroup,
		MaximumAttempts:     config.MaximumAttempts,
		InitialRetryBackoff: config.InitialRetryBackoff,
		MaximumRetryBackoff: config.MaximumRetryBackoff,
		MaximumPollRecords:  config.MaximumPollRecords,
	}
}

func loadKafkaConsumerConfig() (KafkaConsumerConfig, error) {
	connection, err := skafka.LoadConnectionConfig()
	if err != nil {
		return KafkaConsumerConfig{}, err
	}
	consumerGroup := strings.TrimSpace(os.Getenv("NOTIFICATION_KAFKA_CONSUMER_GROUP"))
	if consumerGroup == "" {
		return KafkaConsumerConfig{}, fmt.Errorf("NOTIFICATION_KAFKA_CONSUMER_GROUP is required")
	}
	maximumAttempts, err := strconv.Atoi(strings.TrimSpace(os.Getenv("NOTIFICATION_KAFKA_MAXIMUM_ATTEMPTS")))
	if err != nil || maximumAttempts <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("NOTIFICATION_KAFKA_MAXIMUM_ATTEMPTS must be a positive integer")
	}
	initialRetryBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_KAFKA_INITIAL_RETRY_BACKOFF")))
	if err != nil || initialRetryBackoff <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("NOTIFICATION_KAFKA_INITIAL_RETRY_BACKOFF must be a positive Go duration")
	}
	maximumRetryBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_KAFKA_MAXIMUM_RETRY_BACKOFF")))
	if err != nil || maximumRetryBackoff < initialRetryBackoff {
		return KafkaConsumerConfig{}, fmt.Errorf("NOTIFICATION_KAFKA_MAXIMUM_RETRY_BACKOFF must not be smaller than the initial retry backoff")
	}
	maximumPollRecords, err := strconv.Atoi(strings.TrimSpace(os.Getenv("NOTIFICATION_KAFKA_MAXIMUM_POLL_RECORDS")))
	if err != nil || maximumPollRecords <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("NOTIFICATION_KAFKA_MAXIMUM_POLL_RECORDS must be a positive integer")
	}

	return KafkaConsumerConfig{
		Connection:          connection,
		ConsumerGroup:       consumerGroup,
		MaximumAttempts:     maximumAttempts,
		InitialRetryBackoff: initialRetryBackoff,
		MaximumRetryBackoff: maximumRetryBackoff,
		MaximumPollRecords:  maximumPollRecords,
	}, nil
}
