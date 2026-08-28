package configs

import (
	"fmt"
	"os"
	"strings"
	"time"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Config struct {
	ListenAddress               string
	Postgres                    spostgres.Config
	Kafka                       KafkaConsumerConfig
	NotificationCleanupInterval time.Duration
	NotificationRetention       time.Duration
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
	notificationCleanupInterval, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_CLEANUP_INTERVAL")))
	if err != nil || notificationCleanupInterval <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_CLEANUP_INTERVAL must be a positive Go duration")
	}
	notificationRetention, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_RETENTION")))
	if err != nil || notificationRetention <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_RETENTION must be a positive Go duration")
	}

	return Config{
		ListenAddress:               listenAddress,
		Postgres:                    postgres,
		Kafka:                       kafka,
		NotificationCleanupInterval: notificationCleanupInterval,
		NotificationRetention:       notificationRetention,
	}, nil
}
