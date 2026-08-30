package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	data "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres"
)

var viewAllDatabaseEnumsCommand = &cobra.Command{
	Use:   "viewAllEnums",
	Short: "View all the nums of the database.",
	Long:  "Use a simple select sql command to get all the enums of the database",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := loadCoreDatabaseConfig()
		if err != nil {
			panic(err)
		}
		db, err := spostgres.Connect(config)
		if err != nil {
			panic(err)
		}
		defer spostgres.Disconnect(db)

		if err := spostgres.ViewAllDatabaseEnums(db); err != nil {
			slogs.NotegicLogger.Error(context.Background(), err, "Failed to display database enums")
			return
		}
	},
}

var bootstrapDatabaseCommand = &cobra.Command{
	Use:   "bootstrapDB",
	Short: "Bootstrap the Core PostgreSQL runtime role.",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := loadCoreDatabaseConfig()
		if err != nil {
			panic(err)
		}
		adminConfig, err := loadPostgresAdminConfig()
		if err != nil {
			panic(err)
		}
		adminDB, err := spostgres.Connect(adminConfig)
		if err != nil {
			panic(err)
		}
		defer spostgres.Disconnect(adminDB)

		if err := spostgres.EnsureRuntimeRole(adminDB, ctypes.Runtime_Core, config.Password); err != nil {
			panic(err)
		}
	},
}

var migrateDatabaseCommand = &cobra.Command{
	Use:   "migrateDB",
	Short: "Migrate enums, tables, and some triggers to the database.",
	Long:  "Use some migration SQLs to migrate required enums, tables, and some triggers to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := loadCoreDatabaseConfig()
		if err != nil {
			panic(err)
		}
		adminConfig, err := loadPostgresAdminConfig()
		if err != nil {
			panic(err)
		}
		adminDB, err := spostgres.Connect(adminConfig)
		if err != nil {
			panic(err)
		}
		defer spostgres.Disconnect(adminDB)
		if err := spostgres.EnsureRuntimeRole(adminDB, ctypes.Runtime_Core, config.Password); err != nil {
			panic(err)
		}
		if err := spostgres.EnsureRuntimeRole(
			adminDB,
			ctypes.Runtime_DurableJob,
			os.Getenv("DURABLEJOB_DB_PASSWORD"),
		); err != nil {
			panic(err)
		}
		if err := spostgres.EnsureRuntimeRole(
			adminDB,
			ctypes.Runtime_YjsWorker,
			os.Getenv("YJS_DB_PASSWORD"),
		); err != nil {
			panic(fmt.Errorf("ensure Yjs worker PostgreSQL role: %w", err))
		}

		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of migrating database schema to %v", config.Name))
		if err := spostgres.Migrate(
			adminDB,
			ctypes.Runtime_Core,
			data.DatabaseMigrationManifest,
		); err != nil {
			panic(fmt.Errorf("migrate Core PostgreSQL objects: %w", err))
		}
		if err := spostgres.ApplyPermissions(
			adminDB,
			ctypes.Runtime_Core,
			data.DatabasePermissionManifest,
		); err != nil {
			panic(fmt.Errorf("apply Core PostgreSQL permissions: %w", err))
		}
		if err := spostgres.VerifyPermissions(adminDB, ctypes.Runtime_Core, data.DatabasePermissionManifest); err != nil {
			panic(fmt.Errorf("verify Core PostgreSQL permissions: %w", err))
		}
	},
}

var seedDatabaseCommand = &cobra.Command{
	Use:   "seedDB",
	Short: "Seed some default data for management or main business logic.",
	Long:  "Use some seeding default data SQLs to seed data for management or main business logic.",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := loadCoreDatabaseConfig()
		if err != nil {
			panic(err)
		}
		adminConfig, err := loadPostgresAdminConfig()
		if err != nil {
			panic(err)
		}
		adminDB, err := spostgres.Connect(adminConfig)
		if err != nil {
			panic(err)
		}
		defer spostgres.Disconnect(adminDB)

		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of seeding database default data to %v", config.Name))

		if err := spostgres.Seed(adminDB, ctypes.Runtime_Core, data.DatabaseSeedManifest); err != nil {
			panic(fmt.Errorf("seed Core PostgreSQL data: %w", err))
		}
	},
}

func loadCoreDatabaseConfig() (spostgres.Config, error) {
	return spostgres.LoadConfig(
		os.Getenv("CORE_DB_HOST"),
		os.Getenv("CORE_DB_USER"),
		os.Getenv("CORE_DB_PASSWORD"),
		os.Getenv("CORE_DB_NAME"),
		os.Getenv("CORE_DB_PORT"),
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
