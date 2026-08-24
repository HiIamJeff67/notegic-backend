-- 0000_plan_limitation_seed.down.sql
DELETE FROM "PlanLimitationTable"
WHERE key IN ('Free', 'Pro', 'Premium', 'Ultimate', 'Enterprise');