package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	seeds "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/seeds"
)

var viewAllDatabaseEnumsCommand = &cobra.Command{
	Use:   "viewAllEnums",
	Short: "View all the nums of the database.",
	Long:  "Use a simple select sql command to get all the enums of the database",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := coreconfig.LoadPostgresConfig()
		if err != nil {
			panic(err)
		}
		db, err := data.Connect(config)
		if err != nil {
			panic(err)
		}
		defer data.Disconnect(db)

		if err := spostgres.ViewAllDatabaseEnums(db); err != nil {
			slogs.NotegicLogger.Error(context.Background(), err, "Failed to display database enums")
			return
		}
	},
}

var migrateDatabaseCommand = &cobra.Command{
	Use:   "migrateDB",
	Short: "Migrate enums, tables, and some triggers to the database.",
	Long:  "Use some migration SQLs to migrate required enums, tables, and some triggers to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := coreconfig.LoadPostgresConfig()
		if err != nil {
			panic(err)
		}
		db, err := data.Connect(config)
		if err != nil {
			panic(err)
		}
		defer data.Disconnect(db)

		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of migrating database schema to %v", config.Name))

		for _, migrate := range []func() error{
			func() error {
				return spostgres.Migrate(
					db,
					ctypes.Runtime_Core,
					data.DatabaseMigrationManifest,
				)
			},
		} {
			if err := migrate(); err != nil {
				slogs.NotegicLogger.Error(context.Background(), err, "Failed to migrate database schema")
				return
			}
		}
	},
}

var seedDatabaseCommand = &cobra.Command{
	Use:   "seedDB",
	Short: "Seed some default data for management or main business logic.",
	Long:  "Use some seeding default data SQLs to seed data for management or main business logic.",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := coreconfig.LoadPostgresConfig()
		if err != nil {
			panic(err)
		}
		db, err := data.Connect(config)
		if err != nil {
			panic(err)
		}
		defer data.Disconnect(db)

		slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of seeding database default data to %v", config.Name))

		if err := spostgres.SeedDefaultDataToDatabase(db, seeds.SeedingDefaultDataSQLs); err != nil {
			slogs.NotegicLogger.Error(context.Background(), err, "Failed to seed database default data")
			return
		}
	},
}
