package postgres

import (
	"fmt"
	"strings"
)

type Config struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

// LoadConfig validates explicit PostgreSQL connection values. Runtime config
// packages own environment-variable names and read them before calling here.
func LoadConfig(host, user, password, name, port string) (Config, error) {
	config := Config{
		Host:     strings.TrimSpace(host),
		User:     strings.TrimSpace(user),
		Password: password,
		Name:     strings.TrimSpace(name),
		Port:     strings.TrimSpace(port),
	}
	if config.Host == "" || config.User == "" || config.Password == "" || config.Name == "" || config.Port == "" {
		return Config{}, fmt.Errorf("host, user, password, name, and port are required")
	}

	return config, nil
}
