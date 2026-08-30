package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	sharedstrings "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
)

type Config struct {
	ListenAddress              string
	TrustedProxies             []string
	AllowedDomains             []string
	CoreBaseUrl                string
	CoreAdapterTimeout         time.Duration
	NotificationBaseUrl        string
	NotificationAdapterTimeout time.Duration
}

func LoadConfig() (Config, error) {
	listenAddress := strings.TrimSpace(os.Getenv("CLIENT_GATEWAY_LISTEN_ADDRESS"))
	if listenAddress == "" {
		// Keep the former variable as a migration fallback for existing deployments.
		listenAddress = strings.TrimSpace(os.Getenv("GATEWAY_LISTEN_ADDRESS"))
	}
	config := Config{
		ListenAddress:       listenAddress,
		TrustedProxies:      sharedstrings.SplitValues(os.Getenv("GIN_TRUSTED_PROXIES")),
		AllowedDomains:      sharedstrings.SplitValues(os.Getenv("ALLOWED_DOMAINS")),
		CoreBaseUrl:         strings.TrimRight(strings.TrimSpace(os.Getenv("CORE_BASE_URL")), "/"),
		NotificationBaseUrl: strings.TrimRight(strings.TrimSpace(os.Getenv("NOTIFICATION_BASE_URL")), "/"),
	}
	if config.ListenAddress == "" || config.CoreBaseUrl == "" || config.NotificationBaseUrl == "" {
		return Config{}, fmt.Errorf("CLIENT_GATEWAY_LISTEN_ADDRESS, CORE_BASE_URL, and NOTIFICATION_BASE_URL are required")
	}
	coreTimeout, err := time.ParseDuration(strings.TrimSpace(os.Getenv("CORE_CLIENT_TIMEOUT")))
	if err != nil || coreTimeout <= 0 {
		return Config{}, fmt.Errorf("CORE_CLIENT_TIMEOUT must be a positive Go duration")
	}
	config.CoreAdapterTimeout = coreTimeout
	notificationTimeout, err := time.ParseDuration(strings.TrimSpace(os.Getenv("NOTIFICATION_CLIENT_TIMEOUT")))
	if err != nil || notificationTimeout <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_CLIENT_TIMEOUT must be a positive Go duration")
	}
	config.NotificationAdapterTimeout = notificationTimeout
	return config, nil
}
