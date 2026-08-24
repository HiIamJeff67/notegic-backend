package billingplanseeds

import (
	_ "embed"
)

//go:embed 0000_billing_plan_seed.up.sql
var BillingPlanSeedingDefaultDataSQL_0000_UP string

//go:embed 0000_billing_plan_seed.up.sql
var BillingPlanSeedingDefaultDataSQL_0000_DOWN string
