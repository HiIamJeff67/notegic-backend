DROP INDEX IF EXISTS "users_to_billing_plans_idx_user_id_billing_plan_id_partial_status";

-- ============================== SQL Separator ==============================

CREATE UNIQUE INDEX users_to_billing_plans_idx_user_id_billing_plan_id_partial_status
ON "UsersToBillingPlansTable" (user_id, billing_plan_id)
WHERE status = 'APPROVAL_PENDING' or status = 'APPROVED' or status = 'ACTIVE';