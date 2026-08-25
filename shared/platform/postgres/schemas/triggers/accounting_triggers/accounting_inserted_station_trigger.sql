CREATE OR REPLACE FUNCTION trigger_function_accounting_inserted_station()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
    IF (TG_OP <> 'INSERT') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_inserted_station: %. Expected INSERT.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    FOR r IN
        WITH owner_deltas AS (
            SELECT
                owner_id,
                count(*) as total_delta
            FROM new_table
            GROUP BY owner_id
        ),
        updated_accounts AS (
            UPDATE "UserAccountTable" ua
            SET
                station_count = station_count + od.total_delta,
                updated_at = NOW()
            FROM owner_deltas od
            WHERE ua.user_id = od.owner_id
            RETURNING ua.user_id, ua.station_count
        )

        SELECT
            u.id, u.plan, ua.station_count, pl.max_station_count
        FROM updated_accounts ua
        JOIN "UserTable" u ON ua.user_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE ua.station_count > pl.max_station_count
    LOOP
        RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % stations. Current count: %.',
            r.plan, r.max_station_count, r.station_count
        USING ERRCODE = 'check_violation';
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_inserted_station ON "StationTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_inserted_station
    AFTER INSERT
    ON "StationTable"
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_inserted_station();
