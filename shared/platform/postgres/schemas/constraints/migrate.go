package constraints

import (
	blockconstraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints/block_constraints"
	routineconstraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints/routine_constraints"
	userquotaconstraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints/user_quota_constraints"
	userstobillingplansconstraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints/users_to_billing_plans_constraints"
)

var MigratingConstraintSQLs = []string{
	userstobillingplansconstraints.UserIdBillingPlanIdPartialStatusIndexSQL,
	blockconstraints.BlockSiblingPointerConstraintsSQL,
	routineconstraints.RoutineScheduledTimeMinutePrecisionCheckSQL,
	routineconstraints.RoutineScheduledTimeInPeriodCheckSQL,
}

var UserQuotaConstraintSQLs = []string{
	userquotaconstraints.UserQuotaConstraintSQL,
}
