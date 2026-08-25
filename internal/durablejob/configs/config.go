package config

import (
	"fmt"
	"os"
	"strings"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Config struct {
	ListenAddress             string
	Postgres                  spostgres.Config
	KafkaConsumer             KafkaConsumerConfig
	YjsDocumentInitialization YjsDocumentInitializationConfig
}

func LoadConfig() (Config, error) {
	config := Config{
		ListenAddress: strings.TrimSpace(os.Getenv("DURABLEJOB_LISTEN_ADDRESS")),
	}
	if config.ListenAddress == "" {
		return Config{}, fmt.Errorf("DURABLEJOB_LISTEN_ADDRESS is required")
	}

	postgres, err := loadPostgresConfig()
	if err != nil {
		return Config{}, err
	}
	config.Postgres = postgres

	config.KafkaConsumer, err = loadKafkaConsumerConfig()
	if err != nil {
		return Config{}, err
	}
	config.YjsDocumentInitialization, err = loadYjsDocumentInitializationConfig()
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
