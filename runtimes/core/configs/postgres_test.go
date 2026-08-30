package config

import "testing"

func TestLoadPostgresConfig(t *testing.T) {
	t.Setenv("CORE_DB_HOST", "database")
	t.Setenv("CORE_DB_USER", "notegic_core")
	t.Setenv("CORE_DB_PASSWORD", "secret")
	t.Setenv("CORE_DB_NAME", "notegic")
	t.Setenv("CORE_DB_PORT", "5432")

	config, err := loadPostgresConfig()
	if err != nil {
		t.Fatalf("loadPostgresConfig() error = %v", err)
	}
	if config.Host != "database" || config.Port != "5432" {
		t.Fatalf("loadPostgresConfig() = %#v", config)
	}
}
