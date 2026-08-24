package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

var (
	// DB is the main database instance of the application.
	DB *gorm.DB

	// Maintain the static information about database instances and their config.
	DatabaseInstanceToConfig = map[*gorm.DB]platformpostgres.Config{}
	DatabaseNameToInstance   = map[string]*gorm.DB{}
)

func Connect(config platformpostgres.Config) *gorm.DB {
	dbConn, err := platformpostgres.Connect(config)
	if err != nil {
		logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("Error connecting to the %s database\n", config.Name))
		panic("Connecting to database error : " + err.Error())
	}

	if _, ok := DatabaseInstanceToConfig[dbConn]; !ok {
		logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Storing database of %s into the DatabaseInstanceToConfig...", config.Name))
		DatabaseInstanceToConfig[dbConn] = config
	}
	if _, ok := DatabaseNameToInstance[config.Name]; !ok {
		logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Storing database of %s into the DatabaseNameToInstance...", config.Name))
		DatabaseNameToInstance[config.Name] = dbConn
	}

	logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("%s database connected\n", config.Name))
	return dbConn
}

func Disconnect(db *gorm.DB) bool {
	config, ok := DatabaseInstanceToConfig[db]
	if !ok {
		logs.NotegicLogger.Error(context.Background(), nil, "Failed to get the connection of the given database")
		return false
	}

	if err := platformpostgres.Disconnect(db); err != nil {
		logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("Failed to close the connection of %s database", config.Name))
		return false
	}

	delete(DatabaseInstanceToConfig, db)
	delete(DatabaseNameToInstance, config.Name)
	logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("%s database connection closed", config.Name))
	return true
}
