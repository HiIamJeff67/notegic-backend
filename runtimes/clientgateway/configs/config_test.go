package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("CLIENT_GATEWAY_LISTEN_ADDRESS", "127.0.0.1:7777")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	t.Setenv("NOTIFICATION_BASE_URL", "http://notification:7781")
	t.Setenv("NOTIFICATION_CLIENT_TIMEOUT", "10s")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.CoreAdapterTimeout != 10*time.Second {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}

func TestLoadConfigFallsBackToLegacyGatewayListenAddress(t *testing.T) {
	t.Setenv("CLIENT_GATEWAY_LISTEN_ADDRESS", "")
	t.Setenv("GATEWAY_LISTEN_ADDRESS", "127.0.0.1:7777")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	t.Setenv("NOTIFICATION_BASE_URL", "http://notification:7781")
	t.Setenv("NOTIFICATION_CLIENT_TIMEOUT", "10s")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.ListenAddress != "127.0.0.1:7777" {
		t.Fatalf("LoadConfig() listen address = %q", config.ListenAddress)
	}
}

func TestLoadConfigRequiresClientGatewayListenAddress(t *testing.T) {
	t.Setenv("CLIENT_GATEWAY_LISTEN_ADDRESS", "")
	t.Setenv("GATEWAY_LISTEN_ADDRESS", "")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	t.Setenv("NOTIFICATION_BASE_URL", "http://notification:7781")
	t.Setenv("NOTIFICATION_CLIENT_TIMEOUT", "10s")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() expected missing ClientGateway listen address error")
	}
}
