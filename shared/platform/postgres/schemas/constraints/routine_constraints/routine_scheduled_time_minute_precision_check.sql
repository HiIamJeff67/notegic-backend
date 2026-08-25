ALTER TABLE "RoutineTable" DROP CONSTRAINT IF EXISTS routine_check_scheduled_start_minute_precision;
ALTER TABLE "RoutineTable" DROP CONSTRAINT IF EXISTS routine_check_scheduled_end_minute_precision;

-- ============================== SQL Separator ==============================

ALTER TABLE "RoutineTable"
ADD CONSTRAINT routine_check_scheduled_start_minute_precision
CHECK (
    scheduled_start_at IS NULL
    OR date_trunc('minute', scheduled_start_at) = scheduled_start_at
);

-- ============================== SQL Separator ==============================

ALTER TABLE "RoutineTable"
ADD CONSTRAINT routine_check_scheduled_end_minute_precision
CHECK (
    scheduled_end_at IS NULL
    OR date_trunc('minute', scheduled_end_at) = scheduled_end_at
);
