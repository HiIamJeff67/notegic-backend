DO $$
DECLARE
	constraint_record RECORD;
BEGIN
	FOR constraint_record IN
		SELECT tables.relname AS table_name, constraints.conname
		FROM pg_constraint constraints
		JOIN pg_class tables ON tables.oid = constraints.conrelid
		JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
		WHERE schemas.nspname = 'public'
		  AND constraints.contype = 'f'
		  AND (
			(tables.relname = 'UsersToStationsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (station_id)%')
			OR (tables.relname = 'RoutineTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (station_id)%')
			OR (tables.relname = 'RoutinesToTagsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (routine_id, station_id)%')
			OR (tables.relname = 'RoutinesToTagsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (tag_id, user_id)%')
			OR (tables.relname = 'RoutinesToTagsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id, station_id)%')
			OR (tables.relname = 'RoutinesToItemsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (routine_id)%')
			OR (tables.relname = 'RoutinesToItemsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (item_id, type)%')
		  )
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			constraint_record.table_name,
			constraint_record.conname
		);
	END LOOP;

	CREATE UNIQUE INDEX IF NOT EXISTS "routine_tag_idx_id_owner_id"
		ON "RoutineTagTable" (id, owner_id);

	ALTER TABLE "UsersToStationsTable"
		ADD CONSTRAINT "users_to_stations_station_id_fkey"
		FOREIGN KEY (station_id) REFERENCES "StationTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutineTable"
		ADD CONSTRAINT "routine_station_id_fkey"
		FOREIGN KEY (station_id) REFERENCES "StationTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutinesToTagsTable"
		ADD CONSTRAINT "routines_to_tags_routine_id_station_id_fkey"
		FOREIGN KEY (routine_id, station_id)
		REFERENCES "RoutineTable" (id, station_id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutinesToTagsTable"
		ADD CONSTRAINT "routines_to_tags_tag_id_user_id_fkey"
		FOREIGN KEY (tag_id, user_id)
		REFERENCES "RoutineTagTable" (id, owner_id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutinesToTagsTable"
		ADD CONSTRAINT "routines_to_tags_user_id_station_id_fkey"
		FOREIGN KEY (user_id, station_id)
		REFERENCES "UsersToStationsTable" (user_id, station_id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutinesToItemsTable"
		ADD CONSTRAINT "routines_to_items_routine_id_fkey"
		FOREIGN KEY (routine_id) REFERENCES "RoutineTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutinesToItemsTable"
		ADD CONSTRAINT "routines_to_items_item_id_type_fkey"
		FOREIGN KEY (item_id, type)
		REFERENCES "ItemTable" (id, type)
		ON UPDATE CASCADE ON DELETE CASCADE;
END
$$;
