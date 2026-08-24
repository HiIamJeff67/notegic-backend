CREATE OR REPLACE FUNCTION trigger_function_accounting_inserted_routine()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
    IF (TG_OP <> 'INSERT') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_inserted_routine: %. Expected INSERT.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    FOR r IN
        WITH station_deltas AS (
            SELECT
                station_id,
                count(*) as total_delta
            FROM new_table
            GROUP BY station_id
        ),
        updated_stations AS (
            UPDATE "StationTable" s
            SET
                routine_count = routine_count + sd.total_delta,
                updated_at = NOW()
            FROM station_deltas sd
            WHERE s.id = sd.station_id
            RETURNING s.id, s.owner_id, s.routine_count
        ),
        owner_deltas AS (
            SELECT
                us.owner_id,
                sum(sd.total_delta) as total_delta
            FROM updated_stations us
            JOIN station_deltas sd ON sd.station_id = us.id
            GROUP BY us.owner_id
        ),
        updated_accounts AS (
            UPDATE "UserAccountTable" ua
            SET
                routine_count = routine_count + od.total_delta,
                updated_at = NOW()
            FROM owner_deltas od
            WHERE ua.user_id = od.owner_id
            RETURNING ua.user_id
        )

        SELECT
            us.id, u.plan, us.routine_count, pl.max_routine_count_per_station
        FROM updated_stations us
        JOIN "UserTable" u ON us.owner_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        LEFT JOIN updated_accounts ua ON ua.user_id = us.owner_id
        WHERE us.routine_count > pl.max_routine_count_per_station
    LOOP
        RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % routines per station. Current count: %.',
            r.plan, r.max_routine_count_per_station, r.routine_count
        USING ERRCODE = 'check_violation';
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_inserted_routine ON "RoutineTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_inserted_routine
    AFTER INSERT
    ON "RoutineTable"
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_inserted_routine();
