package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("CORE_LISTEN_ADDRESS", "127.0.0.1:7778")
	t.Setenv("DB_HOST", "database")
	t.Setenv("DB_USER", "notegic")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "notegic")
	t.Setenv("DOCKER_DB_PORT", "5432")
	t.Setenv("OAUTH_GOOGLE_CLIENT_ID", "client-id")
	t.Setenv("OAUTH_GOOGLE_CLIENT_SECRET", "client-secret")
	t.Setenv("OAUTH_GOOGLE_REDIRECT_URL", "http://gateway/auth/google/callback")
	t.Setenv("OUTBOX_RELAY_BATCH_SIZE", "100")
	t.Setenv("OUTBOX_RELAY_POLL_INTERVAL", "1s")
	t.Setenv("OUTBOX_RELAY_CLAIM_TIMEOUT", "30s")
	t.Setenv("OUTBOX_RELAY_INITIAL_BACKOFF", "1s")
	t.Setenv("OUTBOX_RELAY_MAXIMUM_BACKOFF", "1m")
	t.Setenv("OUTBOX_RELAY_RETENTION", "168h")
	t.Setenv("OUTBOX_RELAY_CLEANUP_INTERVAL", "1h")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_ATTEMPTS", "3")
	t.Setenv("KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF", "250ms")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF", "5s")
	t.Setenv("KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS", "100")
	t.Setenv("CORE_QUOTA_CYCLE_WORKER_INTERVAL", "24h")
	t.Setenv("STORAGE_KEY_SALT", "salt")
	t.Setenv("CORE_USER_DATA_CACHE_EXPIRES_IN", "1h")
	t.Setenv("CORE_USER_DATA_CACHE_MAX_ROTATION_RETRIES", "5")
	t.Setenv("YJS_DOCUMENT_INITIALIZATION_WORKER_URL", "http://notegic-yjs-worker:8787/core/yjs-document-initialization/v1")
	t.Setenv("YJS_DOCUMENT_INITIALIZATION_WORKER_TIMEOUT", "30s")
	t.Setenv("CORE_YJS_MAINTENANCE_MAXIMUM_PENDING_HINTS", "1000")
	t.Setenv("CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_BATCH", "32")
	t.Setenv("CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_WORKERS", "8")
	t.Setenv("CORE_YJS_MAINTENANCE_MAXIMUM_REQUEST_ATTEMPTS", "3")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.OutboxRelay.BatchSize != 100 {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}
