CREATE OR REPLACE FUNCTION trigger_function_accounting_mutated_station()
RETURNS TRIGGER AS $$
DECLARE
    target_plan_name TEXT;
    max_station_count INTEGER;
    max_routine_count_per_station INTEGER;
    transferred_station_count BIGINT;
BEGIN
    IF TG_OP <> 'UPDATE' THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_mutated_station: %. Expected UPDATE.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    IF NEW.owner_id IS NOT DISTINCT FROM OLD.owner_id THEN
        RETURN NEW;
    END IF;

    SELECT
        pl.max_station_count,
        pl.max_routine_count_per_station,
        u.plan::TEXT
    INTO
        max_station_count,
        max_routine_count_per_station,
        target_plan_name
    FROM "UserTable" u
    JOIN "PlanLimitationTable" pl ON pl.key = u.plan
    WHERE u.id = NEW.owner_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Data integrity: Cannot find new owner for Station. Possible orphan record.'
        USING ERRCODE = 'data_exception';
    END IF;

    IF NEW.routine_count > max_routine_count_per_station THEN
        RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % routines per station. Current count: %.',
            target_plan_name, max_routine_count_per_station, NEW.routine_count
        USING ERRCODE = 'check_violation';
    END IF;

    UPDATE "UserAccountTable"
    SET
        station_count = GREATEST(0, station_count - 1),
        routine_count = GREATEST(0, routine_count - OLD.routine_count),
        updated_at = NOW()
    WHERE user_id = OLD.owner_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the previous Station owner (Owner ID: %).', OLD.owner_id
        USING ERRCODE = 'integrity_constraint_violation';
    END IF;

    UPDATE "UserAccountTable"
    SET
        station_count = station_count + 1,
        routine_count = routine_count + NEW.routine_count,
        updated_at = NOW()
    WHERE user_id = NEW.owner_id
    RETURNING station_count
    INTO transferred_station_count;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the new Station owner (Owner ID: %).', NEW.owner_id
        USING ERRCODE = 'integrity_constraint_violation';
    END IF;

    IF transferred_station_count > max_station_count THEN
        RAISE EXCEPTION 'Quota exceeded while transferring Station ownership to plan "%".', target_plan_name
        USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_mutated_station ON "StationTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_mutated_station
    BEFORE UPDATE
    ON "StationTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_mutated_station();
