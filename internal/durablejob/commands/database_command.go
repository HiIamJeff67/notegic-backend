package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	data "github.com/HiIamJeff67/notegic-backend/internal/durablejob/data/postgres"
)

var migrateDatabaseCommand = &cobra.Command{
	Use:   "migrateDB",
	Short: "Migrate DurableJob PostgreSQL objects and permissions.",
	Run: func(_ *cobra.Command, _ []string) {
		runtimeConfig, err := loadDurableJobDatabaseConfig()
		if err != nil {
			panic(err)
		}
		adminConfig, err := loadPostgresAdminConfig()
		if err != nil {
			panic(err)
		}
		adminDB, err := spostgres.Connect(adminConfig)
		if err != nil {
			panic(fmt.Errorf("connect DurableJob PostgreSQL admin database: %w", err))
		}
		defer spostgres.Disconnect(adminDB)
		if err := spostgres.EnsureRuntimeRole(adminDB, ctypes.Runtime_DurableJob, runtimeConfig.Password); err != nil {
			panic(fmt.Errorf("bootstrap DurableJob PostgreSQL role: %w", err))
		}
		if err := spostgres.GrantMigrationAccess(adminDB, ctypes.Runtime_DurableJob); err != nil {
			panic(fmt.Errorf("grant DurableJob migration access: %w", err))
		}
		defer spostgres.RevokeMigrationAccess(adminDB, ctypes.Runtime_DurableJob)

		db, err := spostgres.Connect(runtimeConfig)
		if err != nil {
			panic(fmt.Errorf("connect DurableJob PostgreSQL database: %w", err))
		}
		defer spostgres.Disconnect(db)
		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start DurableJob database migration in %v", runtimeConfig.Name))
		if err := spostgres.Migrate(db, ctypes.Runtime_DurableJob, data.DatabaseMigrationManifest); err != nil {
			panic(err)
		}
		if err := spostgres.ApplyPermissions(adminDB, ctypes.Runtime_DurableJob, data.DatabasePermissionManifest); err != nil {
			panic(err)
		}
	},
}

func loadDurableJobDatabaseConfig() (spostgres.Config, error) {
	return spostgres.LoadConfig(
		os.Getenv("DURABLEJOB_DB_HOST"),
		os.Getenv("DURABLEJOB_DB_USER"),
		os.Getenv("DURABLEJOB_DB_PASSWORD"),
		os.Getenv("DURABLEJOB_DB_NAME"),
		os.Getenv("DURABLEJOB_DB_PORT"),
	)
}

func loadPostgresAdminConfig() (spostgres.Config, error) {
	return spostgres.LoadConfig(
		os.Getenv("DB_ADMIN_HOST"),
		os.Getenv("DB_ADMIN_USER"),
		os.Getenv("DB_ADMIN_PASSWORD"),
		os.Getenv("DB_ADMIN_NAME"),
		os.Getenv("DB_ADMIN_PORT"),
	)
}
