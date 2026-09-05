package config

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Setenv("EMAIL_LISTEN_ADDRESS", "127.0.0.1:8081")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("NOTEGIC_OFFICIAL_NAME", "Notegic")
	t.Setenv("NOTEGIC_OFFICIAL_GMAIL", "noreply@example.com")
	t.Setenv("NOTEGIC_OFFICIAL_GOOGLE_APPLICATION_PASSWORD", "secret")
	t.Setenv("NOTEGIC_WEB_BASE_URL", "https://notegic.example.com")
	t.Setenv("NOTEGIC_TERMS_URL", "https://notegic.example.com/terms")
	t.Setenv("KAFKA_BROKERS", "kafka:9092")
	t.Setenv("KAFKA_DIAL_TIMEOUT", "3s")
	t.Setenv("KAFKA_TLS_ENABLED", "false")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.SMTP.Port != 587 || config.SMTP.From != "Notegic <noreply@example.com>" {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}
