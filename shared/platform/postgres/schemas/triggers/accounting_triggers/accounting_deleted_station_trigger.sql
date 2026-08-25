CREATE OR REPLACE FUNCTION trigger_function_accounting_deleted_station()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP <> 'DELETE') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_deleted_station: %. Expected DELETE.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    WITH owner_deltas AS (
        SELECT
            owner_id,
            count(*) as station_delta
        FROM old_table
        GROUP BY owner_id
    )
    UPDATE "UserAccountTable" ua
    SET
        station_count = GREATEST(0, station_count - od.station_delta),
        updated_at = NOW()
    FROM owner_deltas od
    WHERE ua.user_id = od.owner_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_deleted_station ON "StationTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_deleted_station
    AFTER DELETE
    ON "StationTable"
    REFERENCING OLD TABLE AS old_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_deleted_station();
