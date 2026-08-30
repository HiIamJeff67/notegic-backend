package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("API_GATEWAY_LISTEN_ADDRESS", "127.0.0.1:7780")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.CoreAdapterTimeout != 10*time.Second {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}

func TestLoadConfigRequiresAPIGatewayListenAddress(t *testing.T) {
	t.Setenv("API_GATEWAY_LISTEN_ADDRESS", "")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() expected missing API gateway listen address error")
	}
}

func TestLoadConfigUsesAPIGatewayListenAddress(t *testing.T) {
	t.Setenv("API_GATEWAY_LISTEN_ADDRESS", "127.0.0.1:7780")
	t.Setenv("CORE_BASE_URL", "http://core:7778")
	t.Setenv("CORE_CLIENT_TIMEOUT", "10s")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.ListenAddress != "127.0.0.1:7780" {
		t.Fatalf("LoadConfig() listen address = %q", config.ListenAddress)
	}
}
