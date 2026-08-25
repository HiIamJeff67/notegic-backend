package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Config struct {
	ListenAddress         string
	Postgres              platformpostgres.Config
	Kafka                 KafkaConsumerConfig
	OutboxPollInterval    time.Duration
	OutboxClaimTimeout    time.Duration
	OutboxInitialBackoff  time.Duration
	OutboxMaximumBackoff  time.Duration
	OutboxBatchSize       int
	OutboxCleanupInterval time.Duration
	OutboxRetention       time.Duration
	NotificationRetention time.Duration
}

func LoadConfig() (Config, error) {
	listenAddress := strings.TrimSpace(os.Getenv("NOTIFICATION_LISTEN_ADDRESS"))
	if listenAddress == "" {
		return Config{}, fmt.Errorf("NOTIFICATION_LISTEN_ADDRESS is required")
	}
	if strings.TrimSpace(os.Getenv("CORE_DELEGATION_SECRET")) == "" ||
		strings.TrimSpace(os.Getenv("CORE_DELEGATION_AUDIENCE")) == "" ||
		strings.TrimSpace(os.Getenv("CORE_DELEGATION_ISSUER")) == "" {
		return Config{}, fmt.Errorf("CORE_DELEGATION_SECRET, CORE_DELEGATION_AUDIENCE, and CORE_DELEGATION_ISSUER are required")
	}
	postgres, err := loadPostgresConfig()
	if err != nil {
		return Config{}, err
	}
	kafka, err := loadKafkaConsumerConfig()
	if err != nil {
		return Config{}, err
	}
	outboxBatchSize, err := strconv.Atoi(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_BATCH_SIZE")))
	if err != nil || outboxBatchSize <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_BATCH_SIZE must be a positive integer")
	}
	outboxPollInterval, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_POLL_INTERVAL")))
	if err != nil || outboxPollInterval <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_POLL_INTERVAL must be a positive Go duration")
	}
	outboxClaimTimeout, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_CLAIM_TIMEOUT")))
	if err != nil || outboxClaimTimeout <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_CLAIM_TIMEOUT must be a positive Go duration")
	}
	outboxInitialBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_INITIAL_BACKOFF")))
	if err != nil || outboxInitialBackoff <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_INITIAL_BACKOFF must be a positive Go duration")
	}
	outboxMaximumBackoff, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_MAXIMUM_BACKOFF")))
	if err != nil || outboxMaximumBackoff < outboxInitialBackoff {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_MAXIMUM_BACKOFF must be greater than or equal to NOTIFICATION_OUTBOX_INITIAL_BACKOFF")
	}
	outboxCleanupInterval, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_CLEANUP_INTERVAL")))
	if err != nil || outboxCleanupInterval <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_CLEANUP_INTERVAL must be a positive Go duration")
	}
	outboxRetention, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_OUTBOX_RETENTION")))
	if err != nil || outboxRetention <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_OUTBOX_RETENTION must be a positive Go duration")
	}
	notificationRetention, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_RETENTION")))
	if err != nil || notificationRetention <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_RETENTION must be a positive Go duration")
	}

	return Config{
		ListenAddress:         listenAddress,
		Postgres:              postgres,
		Kafka:                 kafka,
		OutboxPollInterval:    outboxPollInterval,
		OutboxClaimTimeout:    outboxClaimTimeout,
		OutboxInitialBackoff:  outboxInitialBackoff,
		OutboxMaximumBackoff:  outboxMaximumBackoff,
		OutboxBatchSize:       outboxBatchSize,
		OutboxCleanupInterval: outboxCleanupInterval,
		OutboxRetention:       outboxRetention,
		NotificationRetention: notificationRetention,
	}, nil
}
