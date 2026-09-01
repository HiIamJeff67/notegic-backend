package userquotaconstraints

import (
	_ "embed"
)

//go:embed user_quota_constraints.sql
var UserQuotaConstraintSQL string
