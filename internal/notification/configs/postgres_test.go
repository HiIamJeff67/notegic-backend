package configs

import "testing"

func TestLoadPostgresConfig(t *testing.T) {
	t.Setenv("NOTIFICATION_DB_HOST", "database")
	t.Setenv("NOTIFICATION_DB_USER", "notegic_notification")
	t.Setenv("NOTIFICATION_DB_PASSWORD", "secret")
	t.Setenv("NOTIFICATION_DB_NAME", "notification")
	t.Setenv("NOTIFICATION_DB_PORT", "5432")

	config, err := loadPostgresConfig()
	if err != nil {
		t.Fatalf("loadPostgresConfig() error = %v", err)
	}
	if config.Host != "database" || config.Port != "5432" {
		t.Fatalf("loadPostgresConfig() = %#v", config)
	}
}
