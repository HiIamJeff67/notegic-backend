-- 0000_billing_plan_seed.down.sql
DELETE FROM "BillingPlanTable"
WHERE key IN ('Free', 'Pro', 'Premium', 'Ultimate', 'Enterprise');