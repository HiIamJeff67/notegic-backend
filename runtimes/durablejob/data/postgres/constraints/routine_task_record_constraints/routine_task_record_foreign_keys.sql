DO $$
DECLARE
	constraint_record RECORD;
BEGIN
	FOR constraint_record IN
		SELECT constraints.conname
		FROM pg_constraint constraints
		JOIN pg_class tables ON tables.oid = constraints.conrelid
		JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
		WHERE schemas.nspname = 'public'
		  AND tables.relname = 'RoutineTaskRecordTable'
		  AND constraints.contype = 'f'
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			'RoutineTaskRecordTable',
			constraint_record.conname
		);
	END LOOP;

	ALTER TABLE "RoutineTaskRecordTable"
		ADD CONSTRAINT "routine_task_record_routine_record_id_fkey"
		FOREIGN KEY (routine_record_id) REFERENCES "RoutineRecordTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutineTaskRecordTable"
		ADD CONSTRAINT "routine_task_record_routine_task_id_fkey"
		FOREIGN KEY (routine_task_id) REFERENCES "RoutineTaskTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;
END
$$;
