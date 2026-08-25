package postgres

import "testing"

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("database", "notegic", "secret", "notegic", "5432")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Host != "database" || config.User != "notegic" || config.Password != "secret" || config.Name != "notegic" || config.Port != "5432" {
		t.Fatalf("LoadConfig() = %#v", config)
	}
}

func TestLoadConfigRequiresEveryValue(t *testing.T) {
	if _, err := LoadConfig("database", "notegic", "", "notegic", "5432"); err == nil {
		t.Fatal("LoadConfig() expected missing password error")
	}
}
