CREATE OR REPLACE FUNCTION trigger_function_accounting_inserted_routine_task()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP <> 'INSERT') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_inserted_routine_task: %. Expected INSERT.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    NEW.cost_unit = (octet_length(COALESCE(NEW.payload::text, ''))::bigint + 1023) / 1024;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_inserted_routine_task ON "RoutineTaskTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_inserted_routine_task
    BEFORE INSERT
    ON "RoutineTaskTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_inserted_routine_task();
