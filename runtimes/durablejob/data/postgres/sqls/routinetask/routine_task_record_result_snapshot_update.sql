UPDATE "RoutineTaskRecordTable" AS routine_task_record
SET result_snapshot = results.result_snapshot,
    updated_at = ?
FROM (VALUES %s) AS results(id, result_snapshot)
WHERE routine_task_record.id = results.id
