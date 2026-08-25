package config

import (
	"fmt"
	"os"
	"strings"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Config struct {
	Postgres                  platformpostgres.Config
	ListenAddress             string
	OAuthGoogle               OAuthGoogleConfig
	OutboxRelay               OutboxRelayConfig
	KafkaConsumer             KafkaConsumerConfig
	QuotaCycleWorker          QuotaCycleWorkerConfig
	UserDataCache             UserDataCacheConfig
	YjsDocumentInitialization YjsDocumentInitializationConfig
	YjsMaintenanceStrategy    YjsMaintenanceStrategyConfig
	StorageKeySalt            string
}

func LoadConfig() (Config, error) {
	postgres, err := loadPostgresConfig()
	if err != nil {
		return Config{}, err
	}
	listenAddress := strings.TrimSpace(os.Getenv("CORE_LISTEN_ADDRESS"))
	if listenAddress == "" {
		return Config{}, fmt.Errorf("CORE_LISTEN_ADDRESS is required")
	}
	oauthGoogle, err := loadOAuthGoogleConfig()
	if err != nil {
		return Config{}, err
	}
	outboxRelay, err := loadOutboxRelayConfig()
	if err != nil {
		return Config{}, err
	}
	kafkaConsumer, err := loadKafkaConsumerConfig()
	if err != nil {
		return Config{}, err
	}
	quotaCycleWorker, err := loadQuotaCycleWorkerConfig()
	if err != nil {
		return Config{}, err
	}
	storageKeySalt := os.Getenv("STORAGE_KEY_SALT")
	if storageKeySalt == "" {
		return Config{}, fmt.Errorf("STORAGE_KEY_SALT is required")
	}
	userDataCache, err := loadUserDataCacheConfig()
	if err != nil {
		return Config{}, err
	}
	yjsDocumentInitialization, err := loadYjsDocumentInitializationConfig()
	if err != nil {
		return Config{}, err
	}
	yjsMaintenanceStrategy, err := loadYjsMaintenanceStrategyConfig()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Postgres:                  postgres,
		ListenAddress:             listenAddress,
		OAuthGoogle:               oauthGoogle,
		OutboxRelay:               outboxRelay,
		KafkaConsumer:             kafkaConsumer,
		QuotaCycleWorker:          quotaCycleWorker,
		UserDataCache:             userDataCache,
		YjsDocumentInitialization: yjsDocumentInitialization,
		YjsMaintenanceStrategy:    yjsMaintenanceStrategy,
		StorageKeySalt:            storageKeySalt,
	}, nil
}
