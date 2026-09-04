package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	constraints "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/constraints"
	triggers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/triggers"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by
// DurableJob. Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = spostgres.MigrationManifest{
	Runtime:     ctypes.Runtime_DurableJob,
	Triggers:    triggers.RoutineTaskTriggerSQLs,
	Constraints: constraints.MigratingConstraintSQLs,
	Tables: []any{
		&sschemas.RoutineRecord{},
		&sschemas.RoutineTask{},
		&sschemas.RoutineTaskDependency{},
		&sschemas.RoutineTaskRecord{},
		&sschemas.UserQuota{},
	},
}
