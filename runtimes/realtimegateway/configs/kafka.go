package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type KafkaConsumerConfig struct {
	MaximumAttempts     int
	InitialRetryBackoff time.Duration
	MaximumRetryBackoff time.Duration
	MaximumPollRecords  int
}

func loadKafkaConsumerConfig() (KafkaConsumerConfig, error) {
	maximumAttempts, err := strconv.Atoi(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS")))
	if err != nil || maximumAttempts <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS must be a positive integer")
	}
	initialRetryBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF")))
	if err != nil || initialRetryBackoff <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF must be a positive Go duration")
	}
	maximumRetryBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF")))
	if err != nil || maximumRetryBackoff <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF must be a positive Go duration")
	}
	if maximumRetryBackoff < initialRetryBackoff {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF must not be smaller than KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF")
	}
	maximumPollRecords, err := strconv.Atoi(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS")))
	if err != nil || maximumPollRecords <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS must be a positive integer")
	}

	return KafkaConsumerConfig{
		MaximumAttempts:     maximumAttempts,
		InitialRetryBackoff: initialRetryBackoff,
		MaximumRetryBackoff: maximumRetryBackoff,
		MaximumPollRecords:  maximumPollRecords,
	}, nil
}
