package usesrtobillingplansconstraints

import (
	_ "embed"
)

var (
	//go:embed user_id_billing_plan_id_partial_status_idx.sql
	UserIdBillingPlanIdPartialStatusIndexSQL string
)
