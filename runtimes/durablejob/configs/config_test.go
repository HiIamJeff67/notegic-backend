package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("DURABLEJOB_LISTEN_ADDRESS", "127.0.0.1:8082")
	t.Setenv("DURABLEJOB_DB_HOST", "localhost")
	t.Setenv("DURABLEJOB_DB_USER", "notegic_durablejob")
	t.Setenv("DURABLEJOB_DB_PASSWORD", "postgres")
	t.Setenv("DURABLEJOB_DB_NAME", "notegic")
	t.Setenv("DURABLEJOB_DB_PORT", "5432")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS", "3")
	t.Setenv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF", "250ms")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF", "5s")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS", "100")
	t.Setenv("YJS_DOCUMENT_INITIALIZATION_WORKER_URL", "http://notegic-yjs-worker:8787/core/yjs-document-initialization/v1")
	t.Setenv("YJS_DOCUMENT_INITIALIZATION_WORKER_TIMEOUT", "30s")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.KafkaConsumer.MaximumAttempts != 3 {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}
