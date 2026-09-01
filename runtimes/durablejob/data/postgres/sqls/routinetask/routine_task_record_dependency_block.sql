WITH RECURSIVE blocked_tasks(routine_record_id, routine_task_id) AS (
    SELECT routine_record_id,
        routine_task_id
    FROM "RoutineTaskRecordTable"
    WHERE id IN ?
    UNION
    SELECT blocked_tasks.routine_record_id,
        dependency.routine_task_id
    FROM blocked_tasks
    INNER JOIN "RoutineDependencyTable" dependency
        ON dependency.previous_routine_task_id = blocked_tasks.routine_task_id
)
UPDATE "RoutineTaskRecordTable" AS routine_task_record
SET status = ?,
    updated_at = ?
FROM blocked_tasks
WHERE routine_task_record.routine_record_id = blocked_tasks.routine_record_id
    AND routine_task_record.routine_task_id = blocked_tasks.routine_task_id
    AND routine_task_record.status IN (?, ?)
