package configs

import (
	"fmt"
	"os"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func loadPostgresConfig() (spostgres.Config, error) {
	config, err := spostgres.LoadConfig(
		os.Getenv("NOTIFICATION_DB_HOST"),
		os.Getenv("NOTIFICATION_DB_USER"),
		os.Getenv("NOTIFICATION_DB_PASSWORD"),
		os.Getenv("NOTIFICATION_DB_NAME"),
		os.Getenv("NOTIFICATION_DB_PORT"),
	)
	if err != nil {
		return spostgres.Config{}, fmt.Errorf("Notification PostgreSQL config requires NOTIFICATION_DB_HOST, NOTIFICATION_DB_USER, NOTIFICATION_DB_PASSWORD, NOTIFICATION_DB_NAME, and NOTIFICATION_DB_PORT: %w", err)
	}
	return config, nil
}
