package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	constraints "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/constraints"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	triggers "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/triggers"
	views "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/views"
	seeds "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/seeds"
)

var viewAllAvailableDatabasesCommand = &cobra.Command{
	Use:   "viewDatabases",
	Short: "View all the available databases.",
	Long:  "Use some map to storing and printing the available databases in the project.",
	Run: func(_ *cobra.Command, _ []string) {
		logs.NotegicLogger.Info(context.Background(), "All available databases:")
		for key, value := range data.DatabaseNameToInstance {
			logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("database name: %v, instance: %v", key, value))
		}
	},
}

var viewAllDatabaseEnumsCommand = &cobra.Command{
	Use:   "viewAllEnums",
	Short: "View all the nums of the database.",
	Long:  "Use a simple select sql command to get all the enums of the database",
	Run: func(_ *cobra.Command, _ []string) {
		config, err := coreconfig.LoadPostgresConfig()
		if err != nil {
			panic(err)
		}
		db := data.Connect(config)
		defer data.Disconnect(db)

		if err := platformpostgres.ViewAllDatabaseEnums(db); err != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to display database enums")
			return
		}
	},
}

var truncateDatabaseCommand = &cobra.Command{
	Use:   "truncate",
	Short: "Truncate an existing table",
	Long:  "Truncate the database table with the given table name",
	Run: func(command *cobra.Command, _ []string) {
		databaseName, err := command.Flags().GetString("database")
		if err != nil {
			logs.NotegicLogger.Error(context.Background(), nil, "The --database flag must be specified")
			return
		}

		tableNameString, err := command.Flags().GetString("table")
		if err != nil {
			logs.NotegicLogger.Error(context.Background(), nil, "The --table flag must be specified")
			return
		}

		tableName, exists := data.ConvertToTableName(tableNameString)
		if !exists {
			logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("The table name of %s is not in the database %s", tableNameString, databaseName))
			return
		}

		db, exists := data.DatabaseNameToInstance[tableNameString]
		if !exists {
			logs.NotegicLogger.Error(context.Background(), nil, "The database instance is not exist")
			return
		}

		logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of truncating database table: %s", tableNameString))
		db = data.Connect(data.DatabaseInstanceToConfig[db])
		defer data.Disconnect(db)

		if err := platformpostgres.TruncateTablesInDatabase(tableName, db); err != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to truncate database table")
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
		db := data.Connect(config)
		defer data.Disconnect(db)

		logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of migrating database schema to %v", config.Name))

		for _, migrate := range []func() error{
			func() error { return platformpostgres.MigrateEnumsToDatabase(db, enums.MigratingEnums) },
			func() error { return platformpostgres.MigrateTablesToDatabase(db, schemas.MigratingTables) },
			func() error { return platformpostgres.MigrateViewsToDatabase(db, views.MigratingViewSQLs) },
			func() error { return platformpostgres.MigrateTriggersToDatabase(db, triggers.MigratingTriggerSQLs) },
			func() error {
				return platformpostgres.MigrateConstraintsToDatabase(db, constraints.MigratingConstraintSQLs)
			},
		} {
			if err := migrate(); err != nil {
				logs.NotegicLogger.Error(context.Background(), err, "Failed to migrate database schema")
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
		db := data.Connect(config)
		defer data.Disconnect(db)

		logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Start the process of seeding database default data to %v", config.Name))

		if err := platformpostgres.SeedDefaultDataToDatabase(db, seeds.SeedingDefaultDataSQLs); err != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to seed database default data")
			return
		}
	},
}
