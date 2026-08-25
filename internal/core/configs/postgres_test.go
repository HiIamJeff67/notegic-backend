package config

import "testing"

func TestLoadPostgresConfig(t *testing.T) {
	t.Setenv("DB_HOST", "database")
	t.Setenv("DB_USER", "notegic")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "notegic")
	t.Setenv("DOCKER_DB_PORT", "5432")

	config, err := loadPostgresConfig()
	if err != nil {
		t.Fatalf("loadPostgresConfig() error = %v", err)
	}
	if config.Host != "database" || config.Port != "5432" {
		t.Fatalf("loadPostgresConfig() = %#v", config)
	}
}
