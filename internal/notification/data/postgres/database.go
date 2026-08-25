package postgres

import (
	"gorm.io/gorm"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// DB is the Notification runtime's default connection pool.
var DB *gorm.DB

// defaultDB is owned by the Notification process and is injected into Notification repositories.
var defaultDB *gorm.DB

// DefaultDB returns the Notification runtime's default connection pool.
func DefaultDB() *gorm.DB {
	return defaultDB
}

// SetDefaultDB sets the Notification runtime's default connection pool.
func SetDefaultDB(db *gorm.DB) {
	defaultDB = db
}

func Connect(config spostgres.Config) (*gorm.DB, error) {
	db, err := spostgres.Connect(config)
	if err != nil {
		return nil, err
	}

	DB = db
	SetDefaultDB(db)
	return db, nil
}

func Disconnect(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	if err := spostgres.Disconnect(db); err != nil {
		return err
	}

	if DB == db {
		DB = nil
	}
	if defaultDB == db {
		SetDefaultDB(nil)
	}
	return nil
}
