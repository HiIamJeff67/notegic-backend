WITH consumption(routine_task_id, user_id, cost_unit, priority, scheduled_at) AS (
    VALUES %s
), eligible_consumption AS (
    SELECT ranked.routine_task_id, ranked.user_id, ranked.cost_unit
    FROM (
        SELECT consumption.*, sum(cost_unit) OVER (
            PARTITION BY user_id ORDER BY priority DESC, scheduled_at ASC, routine_task_id ASC
        ) AS accumulated_cost_unit
        FROM consumption
    ) ranked
    JOIN "UserQuotaTable" user_quota ON user_quota.user_id = ranked.user_id
    JOIN "UserView" user_view ON user_view.id = ranked.user_id
    JOIN "PlanLimitationTable" plan_limitation ON plan_limitation.key = user_view.plan
    WHERE user_quota.routine_task_cost_unit_used + ranked.accumulated_cost_unit <= plan_limitation.max_routine_task_cost_unit_count
), consumption_totals AS (
    SELECT user_id, sum(cost_unit) AS cost_unit FROM eligible_consumption GROUP BY user_id
), consumed AS (
    UPDATE "UserQuotaTable" user_quota
    SET routine_task_cost_unit_used = user_quota.routine_task_cost_unit_used + consumption_totals.cost_unit, updated_at = NOW()
    FROM consumption_totals
    WHERE user_quota.user_id = consumption_totals.user_id
        AND user_quota.routine_task_cost_unit_used + consumption_totals.cost_unit <= (
            SELECT plan_limitation.max_routine_task_cost_unit_count
            FROM "UserView" user_view
            JOIN "PlanLimitationTable" plan_limitation ON plan_limitation.key = user_view.plan
            WHERE user_view.id = consumption_totals.user_id
        )
    RETURNING user_quota.user_id
)
SELECT eligible_consumption.routine_task_id
FROM eligible_consumption
JOIN consumed ON consumed.user_id = eligible_consumption.user_id;
