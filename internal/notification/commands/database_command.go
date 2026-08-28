package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	data "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres"
)

var migrateDatabaseCommand = &cobra.Command{
	Use:   "migrateDB",
	Short: "Migrate Notification PostgreSQL objects and permissions.",
	Run: func(_ *cobra.Command, _ []string) {
		runtimeConfig, err := loadNotificationDatabaseConfig()
		if err != nil {
			panic(err)
		}
		adminConfig, err := loadNotificationAdminConfig()
		if err != nil {
			panic(err)
		}
		adminDB, err := spostgres.Connect(adminConfig)
		if err != nil {
			panic(fmt.Errorf("connect Notification PostgreSQL admin database: %w", err))
		}
		defer spostgres.Disconnect(adminDB)
		if err := spostgres.EnsureRuntimeRole(adminDB, ctypes.Runtime_Notification, runtimeConfig.Password); err != nil {
			panic(fmt.Errorf("bootstrap Notification PostgreSQL role: %w", err))
		}
		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start Notification database migration in %v", runtimeConfig.Name))
		if err := spostgres.Migrate(adminDB, ctypes.Runtime_Notification, data.DatabaseMigrationManifest); err != nil {
			panic(err)
		}
		if err := spostgres.ApplyPermissions(adminDB, ctypes.Runtime_Notification, data.DatabasePermissionManifest); err != nil {
			panic(err)
		}
		if err := spostgres.VerifyPermissions(adminDB, ctypes.Runtime_Notification, data.DatabasePermissionManifest); err != nil {
			panic(fmt.Errorf("verify Notification PostgreSQL permissions: %w", err))
		}
	},
}

func loadNotificationDatabaseConfig() (spostgres.Config, error) {
	return spostgres.LoadConfig(
		os.Getenv("NOTIFICATION_DB_HOST"),
		os.Getenv("NOTIFICATION_DB_USER"),
		os.Getenv("NOTIFICATION_DB_PASSWORD"),
		os.Getenv("NOTIFICATION_DB_NAME"),
		os.Getenv("NOTIFICATION_DB_PORT"),
	)
}

func loadNotificationAdminConfig() (spostgres.Config, error) {
	return spostgres.LoadConfig(
		os.Getenv("NOTIFICATION_DB_ADMIN_HOST"),
		os.Getenv("NOTIFICATION_DB_ADMIN_USER"),
		os.Getenv("NOTIFICATION_DB_ADMIN_PASSWORD"),
		os.Getenv("NOTIFICATION_DB_ADMIN_NAME"),
		os.Getenv("NOTIFICATION_DB_ADMIN_PORT"),
	)
}
