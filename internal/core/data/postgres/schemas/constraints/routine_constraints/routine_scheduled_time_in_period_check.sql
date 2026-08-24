ALTER TABLE "RoutineTable" DROP CONSTRAINT IF EXISTS routine_check_scheduled_time_in_period;

-- ============================== SQL Separator ==============================

ALTER TABLE "RoutineTable"
ADD CONSTRAINT routine_check_scheduled_time_in_period
CHECK (
    scheduled_end_at > scheduled_start_at
    AND (
        period IS NULL
        OR scheduled_end_at <= scheduled_start_at + CASE period
            WHEN 'Daily'::"RoutinePeriod" THEN INTERVAL '1 day'
            WHEN 'Weekly'::"RoutinePeriod" THEN INTERVAL '1 week'
            WHEN 'Monthly'::"RoutinePeriod" THEN INTERVAL '1 month'
        END
    )
);
