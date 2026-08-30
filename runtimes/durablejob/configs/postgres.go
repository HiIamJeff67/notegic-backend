package config

import (
	"fmt"
	"os"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func loadPostgresConfig() (spostgres.Config, error) {
	config, err := spostgres.LoadConfig(
		os.Getenv("DURABLEJOB_DB_HOST"),
		os.Getenv("DURABLEJOB_DB_USER"),
		os.Getenv("DURABLEJOB_DB_PASSWORD"),
		os.Getenv("DURABLEJOB_DB_NAME"),
		os.Getenv("DURABLEJOB_DB_PORT"),
	)
	if err != nil {
		return spostgres.Config{}, fmt.Errorf("DurableJob PostgreSQL config requires DURABLEJOB_DB_HOST, DURABLEJOB_DB_USER, DURABLEJOB_DB_PASSWORD, DURABLEJOB_DB_NAME, and DURABLEJOB_DB_PORT: %w", err)
	}
	return config, nil
}
