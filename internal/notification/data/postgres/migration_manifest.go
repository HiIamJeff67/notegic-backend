package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// Notification. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = spostgres.MigrationManifest{
	Runtime: ctypes.Runtime_Notification,
	Tables: []any{
		&sschemas.Notification{},
		&sschemas.UserProjection{},
	},
}
