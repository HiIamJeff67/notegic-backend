UPDATE "RoutineRecordTable" AS routine_record
SET total_task_count = counts.total_task_count,
    waiting_task_count = counts.waiting_task_count,
    blocked_task_count = counts.blocked_task_count,
    updated_at = ?
FROM (
    SELECT routine_record_id,
        COUNT(*)::integer AS total_task_count,
        COUNT(*) FILTER (WHERE status = 'Waiting')::integer AS waiting_task_count,
        COUNT(*) FILTER (WHERE status = 'Blocked')::integer AS blocked_task_count
    FROM "RoutineTaskRecordTable"
    WHERE routine_record_id IN ?
    GROUP BY routine_record_id
) counts
WHERE routine_record.id = counts.routine_record_id
