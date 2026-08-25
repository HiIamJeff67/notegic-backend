package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sconstraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints"
	striggers "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// DurableJob. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = spostgres.MigrationManifest{
	Runtime:     ctypes.Runtime_DurableJob,
	Triggers:    striggers.RoutineTaskTriggerSQLs,
	Constraints: sconstraints.UserQuotaConstraintSQLs,
	Tables: []any{
		&sschemas.RoutineTask{},
		&sschemas.RoutineTaskRecord{},
		&sschemas.UserQuota{},
	},
}
