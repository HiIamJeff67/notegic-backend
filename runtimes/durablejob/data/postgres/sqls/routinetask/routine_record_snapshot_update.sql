UPDATE "RoutineRecordTable" AS routine_record
SET snapshot = snapshots.snapshot,
    updated_at = ?
FROM (VALUES %s) AS snapshots(id, snapshot)
WHERE routine_record.id = snapshots.id
    AND routine_record.snapshot = '{}'::jsonb
