package config

import (
	"fmt"
	"os"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func loadPostgresConfig() (platformpostgres.Config, error) {
	config, err := platformpostgres.LoadConfig(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DOCKER_DB_PORT"),
	)
	if err != nil {
		return platformpostgres.Config{}, fmt.Errorf("DurableJob PostgreSQL config requires DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, and DOCKER_DB_PORT: %w", err)
	}
	return config, nil
}
