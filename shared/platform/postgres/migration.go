package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
		enumSQL := fmt.Sprintf(`
DO $$
BEGIN
	CREATE TYPE %s AS ENUM (%s);
EXCEPTION
	WHEN duplicate_object THEN NULL;
END
$$;`, quoteIdentifier(name), strings.Join(quotedValues, ", "))
		if err := db.Exec(enumSQL).Error; err != nil {
			logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to create enum %s", name))
			return err
		}

		var dbValues []string
		if err := db.Raw(`
			SELECT enumlabel
			FROM pg_enum enums
			JOIN pg_type types ON types.oid = enums.enumtypid
			JOIN pg_namespace schemas ON schemas.oid = types.typnamespace
			WHERE types.typname = ? AND schemas.nspname = 'public'
			ORDER BY enumsortorder;
		`, name).Scan(&dbValues).Error; err != nil {
			logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to get enum %s values", name))
			return err
		}

		for _, value := range values {
			if slices.Contains(dbValues, value) {
				continue
			}
			quotedValue := strings.ReplaceAll(value, "'", "''")
			if err := db.Exec(fmt.Sprintf("ALTER TYPE %s ADD VALUE IF NOT EXISTS '%s';", quoteIdentifier(name), quotedValue)).Error; err != nil {
				logs.NotegicLogger.Error(context.Background(), err, fmt.Sprintf("Failed to add value %q to enum %s", value, name))
				return err
			}
		}

		for _, dbValue := range dbValues {
			if !slices.Contains(values, dbValue) {
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
	// Relations are runtime query metadata, not migration ownership. Migrating
	// with them enabled makes GORM recursively discover tables owned by another
	// runtime (for example, DurableJob can discover Core's User table through
	// RoutineTask.ActorUser). Foreign keys and other cross-runtime constraints
	// are managed explicitly by the manifest instead.
	migrationDB := db.Session(&gorm.Session{NewDB: true})
	migrationDB.IgnoreRelationshipsWhenMigrating = true
	for _, table := range migratingTables {
		if err := migrationDB.AutoMigrate(table); err != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to migrate table")
			return err
		}
	}

	logs.NotegicLogger.Info(context.Background(), "Migration of tables is done")
	return nil
}

// Migrate applies only the manifest for the requested runtime. Each migration
// phase grants schema CREATE, assumes the owner role, and revokes CREATE in
// the same transaction, so a failed or interrupted phase cannot retain access.
func Migrate(db *gorm.DB, runtime ctypes.Runtime, manifest MigrationManifest) error {
	if db == nil {
		return errors.New("admin database connection is required")
	}
	if !manifest.IsFor(runtime) {
		return fmt.Errorf("runtime %q cannot migrate manifest for runtime %q", runtime, manifest.Runtime)
	}

	return db.Connection(func(connection *gorm.DB) error {
		if err := migrateAsRuntime(connection, runtime, func(tx *gorm.DB) error {
			return MigrateEnumsToDatabase(tx, manifest.Enums)
		}); err != nil {
			return err
		}
		// ApplyPermissions runs after the complete migration, so an existing
		// owned table may have had the owner's ACL entry reconciled away by a
		// previous run. Restore privileges only for tables whose PostgreSQL owner
		// is this runtime before SET ROLE is used to recreate their triggers and
		// constraints. A table owned by another runtime is never escalated here.
		for _, table := range manifest.Tables {
			tableNamer, ok := table.(interface{ TableName() string })
			if !ok {
				continue
			}
			var ownedByRuntime bool
			if err := connection.Raw(`
				SELECT EXISTS (
					SELECT 1
					FROM pg_class tables
					JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
					JOIN pg_roles owners ON owners.oid = tables.relowner
					WHERE schemas.nspname = 'public'
					  AND tables.relname = ?
					  AND owners.rolname = ?
				)
			`, tableNamer.TableName(), runtime.RoleName()).Scan(&ownedByRuntime).Error; err != nil {
				return err
			}
			if !ownedByRuntime {
				continue
			}
			if err := connection.Exec(
				"GRANT ALL PRIVILEGES ON TABLE " + quoteIdentifier(tableNamer.TableName()) + " TO " + quoteIdentifier(runtime.RoleName()),
			).Error; err != nil {
				return err
			}
		}
		if err := migrateAsRuntime(connection, runtime, func(tx *gorm.DB) error {
			return MigrateTablesToDatabase(tx, manifest.Tables)
		}); err != nil {
			return err
		}
		if err := migrateAsRuntime(connection, runtime, func(tx *gorm.DB) error {
			return MigrateViewsToDatabase(tx, manifest.Views)
		}); err != nil {
			return err
		}
		if err := migrateAsRuntime(connection, runtime, func(tx *gorm.DB) error {
			return MigrateTriggersToDatabase(tx, manifest.Triggers)
		}); err != nil {
			return err
		}
		return migrateAsRuntime(connection, runtime, func(tx *gorm.DB) error {
			return MigrateConstraintsToDatabase(tx, manifest.Constraints)
		})
	})
}

func migrateAsRuntime(db *gorm.DB, runtime ctypes.Runtime, migrate func(*gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		quotedRoleName := quoteIdentifier(runtime.RoleName())
		if err = tx.Exec("GRANT CREATE ON SCHEMA public TO " + quotedRoleName).Error; err != nil {
			return err
		}
		if err = tx.Exec("SET LOCAL ROLE " + quotedRoleName).Error; err != nil {
			return err
		}

		if err = migrate(tx); err != nil {
			return err
		}
		if err = tx.Exec("RESET ROLE").Error; err != nil {
			return err
		}
		if err = tx.Exec("REVOKE CREATE ON SCHEMA public FROM " + quotedRoleName).Error; err != nil {
			return err
		}

		return nil
	})
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
