package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	sharedstrings "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
)

type Config struct {
	ListenAddress     string
	TrustedProxies    []string
	AllowedDomains    []string
	RealtimeEnabled   bool
	BetaUserPublicIds []string
	YjsWorkerUrls     []string
	KafkaConsumer     KafkaConsumerConfig
}

func LoadConfig() (Config, error) {
	realtimeEnabled, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("REALTIME_ENABLED")))
	if err != nil {
		return Config{}, fmt.Errorf("REALTIME_ENABLED must be a boolean")
	}
	config := Config{
		ListenAddress:     strings.TrimSpace(os.Getenv("REALTIME_GATEWAY_LISTEN_ADDRESS")),
		TrustedProxies:    sharedstrings.SplitValues(os.Getenv("GIN_TRUSTED_PROXIES")),
		AllowedDomains:    sharedstrings.SplitValues(os.Getenv("ALLOWED_DOMAINS")),
		RealtimeEnabled:   realtimeEnabled,
		BetaUserPublicIds: sharedstrings.SplitValues(os.Getenv("REALTIME_BETA_USER_PUBLIC_IDS")),
		YjsWorkerUrls:     sharedstrings.SplitValues(os.Getenv("YJS_WORKER_URLS")),
	}
	if config.ListenAddress == "" || len(config.YjsWorkerUrls) == 0 {
		return Config{}, fmt.Errorf("REALTIME_GATEWAY_LISTEN_ADDRESS and YJS_WORKER_URLS are required")
	}
	config.KafkaConsumer, err = loadKafkaConsumerConfig()
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
