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
		  AND tables.relname = 'RoutineDependencyTable'
		  AND constraints.contype = 'f'
		  AND (
			pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (routine_task_id)%'
			OR pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (previous_routine_task_id)%'
		  )
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			'RoutineDependencyTable',
			constraint_record.conname
		);
	END LOOP;

	ALTER TABLE "RoutineDependencyTable"
		ADD CONSTRAINT "routine_dependency_routine_task_id_fkey"
		FOREIGN KEY (routine_task_id)
		REFERENCES "RoutineTaskTable" (id)
		ON UPDATE CASCADE
		ON DELETE CASCADE;

	ALTER TABLE "RoutineDependencyTable"
		ADD CONSTRAINT "routine_dependency_previous_routine_task_id_fkey"
		FOREIGN KEY (previous_routine_task_id)
		REFERENCES "RoutineTaskTable" (id)
		ON UPDATE CASCADE
		ON DELETE CASCADE;
END
$$;
