DO $$
DECLARE
	constraint_record RECORD;
	has_deferred_constraint BOOLEAN := FALSE;
BEGIN
	FOR constraint_record IN
		SELECT constraints.conname, constraints.condeferrable, constraints.condeferred
		FROM pg_constraint constraints
		JOIN pg_class tables ON tables.oid = constraints.conrelid
		JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
		WHERE schemas.nspname = 'public'
		  AND tables.relname = 'SubShelfTable'
		  AND constraints.contype = 'f'
		  AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (prev_sub_shelf_id)%'
	LOOP
		IF constraint_record.condeferrable AND constraint_record.condeferred THEN
			has_deferred_constraint := TRUE;
			CONTINUE;
		END IF;

		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			'SubShelfTable',
			constraint_record.conname
		);
	END LOOP;

	IF NOT has_deferred_constraint THEN
		ALTER TABLE "SubShelfTable"
			ADD CONSTRAINT "sub_shelf_prev_sub_shelf_id_fkey"
			FOREIGN KEY (prev_sub_shelf_id)
			REFERENCES "SubShelfTable" (id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
			DEFERRABLE INITIALLY DEFERRED;
	END IF;
END
$$;
