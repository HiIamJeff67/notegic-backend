package postgres

import (
	types "github.com/HiIamJeff67/notegic-backend/contracts/types"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	constraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints"
	triggers "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// DurableJob. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = platformpostgres.MigrationManifest{
	Runtime:     types.Runtime_DurableJob,
	Triggers:    triggers.RoutineTaskTriggerSQLs,
	Constraints: constraints.UserQuotaConstraintSQLs,
	Tables: []any{
		&platformschemas.RoutineTask{},
		&platformschemas.RoutineTaskRecord{},
		&platformschemas.UserQuota{},
	},
}
