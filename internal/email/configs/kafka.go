package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type KafkaConnectionConfig = skafka.ConnectionConfig
type KafkaConsumerConfig = skafka.ConsumerConfig

func loadKafkaConfig() (KafkaConnectionConfig, KafkaConsumerConfig, error) {
	kafka, err := skafka.LoadConnectionConfig()
	if err != nil {
		return KafkaConnectionConfig{}, KafkaConsumerConfig{}, err
	}
	maximumAttempts, err := positiveIntEnv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS", 3)
	if err != nil {
		return KafkaConnectionConfig{}, KafkaConsumerConfig{}, err
	}
	initialRetryBackoff, err := positiveDurationEnv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF", time.Second)
	if err != nil {
		return KafkaConnectionConfig{}, KafkaConsumerConfig{}, err
	}
	maximumRetryBackoff, err := positiveDurationEnv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF", 30*time.Second)
	if err != nil {
		return KafkaConnectionConfig{}, KafkaConsumerConfig{}, err
	}
	maximumPollRecords, err := positiveIntEnv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS", 100)
	if err != nil {
		return KafkaConnectionConfig{}, KafkaConsumerConfig{}, err
	}

	return kafka, skafka.ConsumerConfig{
		ClientConfig: skafka.ClientConfig{
			ConnectionConfig: kafka,
			ClientId:         "notegic-email",
		},
		ConsumerGroup:       "notegic-email-core-v1",
		MaximumAttempts:     maximumAttempts,
		InitialRetryBackoff: initialRetryBackoff,
		MaximumRetryBackoff: maximumRetryBackoff,
		MaximumPollRecords:  maximumPollRecords,
	}, nil
}

func positiveIntEnv(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}

	return parsed, nil
}

func positiveDurationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}

	return parsed, nil
}
