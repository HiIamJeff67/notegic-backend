package postgres

import (
	types "github.com/HiIamJeff67/notegic-backend/contracts/types"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// Notification. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = platformpostgres.MigrationManifest{
	Runtime: types.Runtime_Notification,
	Tables: []any{
		&platformschemas.Notification{},
	},
}
