package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres/schemas"
)

func Connect(config platformpostgres.Config) (*gorm.DB, error) {
	db, err := platformpostgres.Connect(config)
	if err != nil {
		return nil, err
	}
	if err := platformpostgres.MigrateTablesToDatabase(db, schemas.MigratingTables); err != nil {
		_ = platformpostgres.Disconnect(db)
		return nil, fmt.Errorf("migrate notification database: %w", err)
	}

	return db, nil
}

func Disconnect(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if err := platformpostgres.Disconnect(db); err != nil {
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(context.Background(), err, "Failed to close Notification database")
		}
		return err
	}

	return nil
}
