UPDATE "RoutineRecordTable" AS routine_record
SET success_task_count = counts.success_task_count,
    failed_task_count = counts.failed_task_count,
    blocked_task_count = counts.blocked_task_count,
    running_task_count = counts.running_task_count,
    waiting_task_count = counts.waiting_task_count,
    status = CASE
        WHEN counts.running_task_count > 0 OR counts.waiting_task_count > 0 THEN ?::"RoutineRecordStatus"
        WHEN counts.failed_task_count = 0 AND counts.blocked_task_count > 0 THEN ?::"RoutineRecordStatus"
        WHEN counts.failed_task_count > 0 THEN ?::"RoutineRecordStatus"
        ELSE ?::"RoutineRecordStatus"
    END,
    actual_ended_at = CASE
        WHEN counts.running_task_count = 0 AND counts.waiting_task_count = 0 THEN ?
        ELSE routine_record.actual_ended_at
    END,
    updated_at = ?
FROM (
    SELECT routine_record_id,
        COUNT(*) FILTER (WHERE status = 'Success')::integer AS success_task_count,
        COUNT(*) FILTER (WHERE status = 'Failed')::integer AS failed_task_count,
        COUNT(*) FILTER (WHERE status = 'Blocked')::integer AS blocked_task_count,
        COUNT(*) FILTER (WHERE status = 'Running')::integer AS running_task_count,
        COUNT(*) FILTER (WHERE status IN ('Waiting', 'Ready'))::integer AS waiting_task_count
    FROM "RoutineTaskRecordTable"
    WHERE routine_record_id IN ?
    GROUP BY routine_record_id
) counts
WHERE routine_record.id = counts.routine_record_id
