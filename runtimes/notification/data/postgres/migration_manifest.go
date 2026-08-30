package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// Notification. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = spostgres.MigrationManifest{
	Runtime: ctypes.Runtime_Notification,
	Enums: map[string][]string{
		new(cenums.UserPlan).Name():   cenums.AllUserPlanStrings,
		new(cenums.UserStatus).Name(): cenums.AllUserStatusStrings,
	},
	Tables: []any{
		&sschemas.InboxEvent{},
		&sschemas.Notification{},
		&sschemas.OutboxEvent{},
		&sschemas.UserProjection{},
	},
}
