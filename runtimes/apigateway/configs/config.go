package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	sharedstrings "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
)

type Config struct {
	ListenAddress      string
	TrustedProxies     []string
	AllowedDomains     []string
	CoreBaseUrl        string
	CoreAdapterTimeout time.Duration
}

func LoadConfig() (Config, error) {
	listenAddress := strings.TrimSpace(os.Getenv("API_GATEWAY_LISTEN_ADDRESS"))
	config := Config{
		ListenAddress:  listenAddress,
		TrustedProxies: sharedstrings.SplitValues(os.Getenv("GIN_TRUSTED_PROXIES")),
		AllowedDomains: sharedstrings.SplitValues(os.Getenv("ALLOWED_DOMAINS")),
		CoreBaseUrl:    strings.TrimRight(strings.TrimSpace(os.Getenv("CORE_BASE_URL")), "/"),
	}
	if config.ListenAddress == "" || config.CoreBaseUrl == "" {
		return Config{}, fmt.Errorf("API_GATEWAY_LISTEN_ADDRESS and CORE_BASE_URL are required")
	}
	coreTimeout, err := time.ParseDuration(strings.TrimSpace(os.Getenv("CORE_CLIENT_TIMEOUT")))
	if err != nil || coreTimeout <= 0 {
		return Config{}, fmt.Errorf("CORE_CLIENT_TIMEOUT must be a positive Go duration")
	}
	config.CoreAdapterTimeout = coreTimeout
	return config, nil
}
