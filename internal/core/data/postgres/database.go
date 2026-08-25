package postgres

import (
	"gorm.io/gorm"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// DB is the Core runtime's default connection pool.
var DB *gorm.DB

func Connect(config platformpostgres.Config) (*gorm.DB, error) {
	db, err := platformpostgres.Connect(config)
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

func Disconnect(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	if err := platformpostgres.Disconnect(db); err != nil {
		return err
	}

	if DB == db {
		DB = nil
	}
	return nil
}
