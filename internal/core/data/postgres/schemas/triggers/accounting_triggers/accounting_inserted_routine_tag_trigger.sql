CREATE OR REPLACE FUNCTION trigger_function_accounting_inserted_routine_tag()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
    IF (TG_OP <> 'INSERT') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_inserted_routine_tag: %. Expected INSERT.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    FOR r IN
        WITH owner_tag_deltas AS (
            SELECT
                owner_id AS user_id,
                count(*) as total_delta
            FROM new_table
            GROUP BY owner_id
        ),
        updated_accounts AS (
            UPDATE "UserAccountTable" ua
            SET
                routine_tag_count = routine_tag_count + otd.total_delta,
                updated_at = NOW()
            FROM owner_tag_deltas otd
            WHERE ua.user_id = otd.user_id
            RETURNING ua.user_id
        )

        SELECT
            ua.user_id,
            u.plan,
            acc.routine_tag_count,
            pl.max_routine_tag_count
        FROM updated_accounts ua
        JOIN "UserAccountTable" acc ON acc.user_id = ua.user_id
        JOIN "UserTable" u ON u.id = ua.user_id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE acc.routine_tag_count > pl.max_routine_tag_count
    LOOP
        RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % routine tags. Current count: %.',
            r.plan, r.max_routine_tag_count, r.routine_tag_count
        USING ERRCODE = 'check_violation';
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_inserted_routine_tag ON "RoutineTagTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_inserted_routine_tag
    AFTER INSERT
    ON "RoutineTagTable"
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_inserted_routine_tag();
