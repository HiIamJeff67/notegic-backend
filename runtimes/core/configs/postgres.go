package config

import (
	"fmt"
	"os"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func loadPostgresConfig() (spostgres.Config, error) {
	config, err := spostgres.LoadConfig(
		os.Getenv("CORE_DB_HOST"),
		os.Getenv("CORE_DB_USER"),
		os.Getenv("CORE_DB_PASSWORD"),
		os.Getenv("CORE_DB_NAME"),
		os.Getenv("CORE_DB_PORT"),
	)
	if err != nil {
		return spostgres.Config{}, fmt.Errorf("Core PostgreSQL config requires CORE_DB_HOST, CORE_DB_USER, CORE_DB_PASSWORD, CORE_DB_NAME, and CORE_DB_PORT: %w", err)
	}
	return config, nil
}
