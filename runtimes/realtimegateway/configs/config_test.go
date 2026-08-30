package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("REALTIME_ENABLED", "true")
	t.Setenv("REALTIME_GATEWAY_LISTEN_ADDRESS", "127.0.0.1:7779")
	t.Setenv("YJS_WORKER_URLS", "ws://yjs:8787/core/realtime/v1")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS", "3")
	t.Setenv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF", "250ms")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF", "5s")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS", "100")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.KafkaConsumer.InitialRetryBackoff != 250*time.Millisecond {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}
