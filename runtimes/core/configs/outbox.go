package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type OutboxRelayConfig struct {
	BatchSize       int
	PollInterval    time.Duration
	ClaimTimeout    time.Duration
	InitialBackoff  time.Duration
	MaximumBackoff  time.Duration
	Retention       time.Duration
	CleanupInterval time.Duration
}

func loadOutboxRelayConfig() (OutboxRelayConfig, error) {
	config := OutboxRelayConfig{}
	batchSize, err := strconv.Atoi(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_BATCH_SIZE")))
	if err != nil || batchSize <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_BATCH_SIZE must be a positive integer")
	}
	config.BatchSize = batchSize
	config.PollInterval, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_POLL_INTERVAL")))
	if err != nil || config.PollInterval <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_POLL_INTERVAL must be a positive Go duration")
	}
	config.ClaimTimeout, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_CLAIM_TIMEOUT")))
	if err != nil || config.ClaimTimeout <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_CLAIM_TIMEOUT must be a positive Go duration")
	}
	config.InitialBackoff, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_INITIAL_BACKOFF")))
	if err != nil || config.InitialBackoff <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_INITIAL_BACKOFF must be a positive Go duration")
	}
	config.MaximumBackoff, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_MAXIMUM_BACKOFF")))
	if err != nil || config.MaximumBackoff <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_MAXIMUM_BACKOFF must be a positive Go duration")
	}
	if config.MaximumBackoff < config.InitialBackoff {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_MAXIMUM_BACKOFF must not be smaller than OUTBOX_RELAY_INITIAL_BACKOFF")
	}
	config.Retention, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_RETENTION")))
	if err != nil || config.Retention <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_RETENTION must be a positive Go duration")
	}
	config.CleanupInterval, err = time.ParseDuration(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_CLEANUP_INTERVAL")))
	if err != nil || config.CleanupInterval <= 0 {
		return OutboxRelayConfig{}, fmt.Errorf("OUTBOX_RELAY_CLEANUP_INTERVAL must be a positive Go duration")
	}

	return config, nil
}
