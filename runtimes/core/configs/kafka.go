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
	config := KafkaConsumerConfig{}
	var err error
	config.MaximumAttempts, err = strconv.Atoi(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS")))
	if err != nil || config.MaximumAttempts <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS must be a positive integer")
	}
	config.InitialRetryBackoff, err = time.ParseDuration(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF")))
	if err != nil || config.InitialRetryBackoff <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF must be a positive Go duration")
	}
	config.MaximumRetryBackoff, err = time.ParseDuration(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF")))
	if err != nil || config.MaximumRetryBackoff <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF must be a positive Go duration")
	}
	if config.MaximumRetryBackoff < config.InitialRetryBackoff {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF must not be smaller than KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF")
	}
	config.MaximumPollRecords, err = strconv.Atoi(strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS")))
	if err != nil || config.MaximumPollRecords <= 0 {
		return KafkaConsumerConfig{}, fmt.Errorf("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS must be a positive integer")
	}

	return config, nil
}
