CREATE OR REPLACE FUNCTION trigger_function_accounting_updated_routine_task()
RETURNS TRIGGER AS $$
DECLARE
    new_cost_unit bigint;
BEGIN
    IF (TG_OP <> 'UPDATE') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_updated_routine_task: %. Expected UPDATE.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    new_cost_unit = (octet_length(COALESCE(NEW.payload::text, ''))::bigint + 1023) / 1024;
    NEW.cost_unit = new_cost_unit;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_updated_routine_task ON "RoutineTaskTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_updated_routine_task
    BEFORE UPDATE OF routine_id, payload, cost_unit
    ON "RoutineTaskTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_updated_routine_task();
