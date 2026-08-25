package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

func ViewAllDatabaseEnums(db *gorm.DB) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	var enumInfos []struct {
		Name   string `gorm:"column:enum_name;"`
		Values string `gorm:"column:enum_values;"`
	}
	result := db.Raw(`
		SELECT
			n.nspname || '.' || t.typname AS enum_name,
			string_agg(e.enumlabel, ', ' ORDER BY e.enumsortorder) AS enum_values
		FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		JOIN pg_namespace n ON n.oid = t.typnamespace
		GROUP BY n.nspname, t.typname
		ORDER BY n.nspname, t.typname;
	`).Scan(&enumInfos)
	if result.Error != nil {
		logs.NotegicLogger.Error(context.Background(), result.Error, "Failed to display database enums")
		return result.Error
	}

	logs.NotegicLogger.Info(context.Background(), "=============== Database Enum List ===============")
	if len(enumInfos) == 0 {
		logs.NotegicLogger.Info(context.Background(), "No enums found")
	} else {
		for index, enumInfo := range enumInfos {
			logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("%d. Type: %-30s | Values: %s", index+1, enumInfo.Name, enumInfo.Values))
		}
	}
	logs.NotegicLogger.Info(context.Background(), "=============== Database Enum List ===============")
	return nil
}

func MigrateEnumsToDatabase(db *gorm.DB, migratingEnums map[string][]string) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	logs.NotegicLogger.Info(context.Background(), "Migrating database enums...")
	for name, values := range migratingEnums {
		quotedValues := make([]string, len(values))
		for index, value := range values {
			quotedValues[index] = "'" + strings.ReplaceAll(value, "'", "''") + "'"
		}
		enumSQL := fmt.Sprintf("CREATE TYPE IF NOT EXISTS %q AS ENUM (%s);", name, strings.Join(quotedValues, ", "))
		if err := db.Exec(enumSQL).Error; err != nil {
			logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to create enum %s", name))
			return err
		}

		var dbValues []string
		if err := db.Raw(`
			SELECT enumlabel
			FROM pg_enum
			WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = ?)
			ORDER BY enumsortorder;
		`, name).Scan(&dbValues).Error; err != nil {
			logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to get enum %s values", name))
			return err
		}

		for _, value := range values {
			if containsString(dbValues, value) {
				continue
			}
			quotedValue := strings.ReplaceAll(value, "'", "''")
			if err := db.Exec(fmt.Sprintf("ALTER TYPE %q ADD VALUE IF NOT EXISTS '%s';", name, quotedValue)).Error; err != nil {
				logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to add value %q to enum %s", value, name))
				return err
			}
		}

		for _, dbValue := range dbValues {
			if !containsString(values, dbValue) {
				logs.NotegicLogger.Warn(context.Background(), fmt.Sprintf("Enum %s contains value %q that is not present in code", name, dbValue))
			}
		}
	}

	logs.NotegicLogger.Info(context.Background(), "Migration of enums is done")
	return nil
}

func MigrateTablesToDatabase(db *gorm.DB, migratingTables []any) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	logs.NotegicLogger.Info(context.Background(), "Migrating database tables...")
	for _, table := range migratingTables {
		if err := db.AutoMigrate(table); err != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to migrate table")
			return err
		}
	}

	logs.NotegicLogger.Info(context.Background(), "Migration of tables is done")
	return nil
}

// Migrate applies only the manifest for the requested runtime. Database
// permissions remain a separate deployment concern.
func Migrate(db *gorm.DB, runtime ctypes.Runtime, manifest MigrationManifest) error {
	if !manifest.IsFor(runtime) {
		return fmt.Errorf("runtime %q cannot migrate manifest for runtime %q", runtime, manifest.Runtime)
	}

	for _, migrate := range []func() error{
		func() error { return MigrateEnumsToDatabase(db, manifest.Enums) },
		func() error { return MigrateTablesToDatabase(db, manifest.Tables) },
		func() error { return MigrateViewsToDatabase(db, manifest.Views) },
		func() error { return MigrateTriggersToDatabase(db, manifest.Triggers) },
		func() error { return MigrateConstraintsToDatabase(db, manifest.Constraints) },
	} {
		if err := migrate(); err != nil {
			return err
		}
	}

	return nil
}

func MigrateTriggersToDatabase(db *gorm.DB, migratingSQLs []string) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	return migrateSQL(db, migratingSQLs, "triggers", true)
}

func MigrateConstraintsToDatabase(db *gorm.DB, migratingSQLs []string) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	return migrateSQL(db, migratingSQLs, "constraints", false)
}

func MigrateViewsToDatabase(db *gorm.DB, migratingSQLs []string) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	return migrateSQL(db, migratingSQLs, "views", false)
}

func SeedDefaultDataToDatabase(db *gorm.DB, seedingSQLs []string) error {
	if logs.NotegicLogger == nil {
		return errors.New("observability logger is not initialized")
	}

	return migrateSQL(db, seedingSQLs, "default data", false)
}

func migrateSQL(db *gorm.DB, sqls []string, description string, skipAlreadyExists bool) error {
	logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Migrating database %s...", description))

	for _, sql := range sqls {
		for _, statement := range strings.Split(sql, SQLSeparator) {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if err := db.Exec(statement).Error; err != nil {
				if skipAlreadyExists && strings.Contains(err.Error(), "SQLSTATE 42710") {
					logs.NotegicLogger.Warn(context.Background(), fmt.Sprintf("Database %s object already exists; skipping: %v", description, err))
					continue
				}
				logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to execute %s SQL statement", description))
				return err
			}
		}
	}

	logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Migration of %s is done", description))
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
