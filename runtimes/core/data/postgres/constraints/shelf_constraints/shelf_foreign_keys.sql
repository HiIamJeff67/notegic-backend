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
			(tables.relname = 'UsersToShelvesTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (root_shelf_id)%')
			OR (tables.relname = 'SubShelfTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (root_shelf_id)%')
			OR (tables.relname = 'MaterialTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (parent_sub_shelf_id)%')
			OR (tables.relname = 'BlockPackTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (parent_sub_shelf_id)%')
			OR (tables.relname = 'ItemTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (parent_sub_shelf_id)%')
			OR (tables.relname = 'ItemTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (root_shelf_id)%')
		  )
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			constraint_record.table_name,
			constraint_record.conname
		);
	END LOOP;

	ALTER TABLE "UsersToShelvesTable"
		ADD CONSTRAINT "users_to_shelves_root_shelf_id_fkey"
		FOREIGN KEY (root_shelf_id) REFERENCES "RootShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "SubShelfTable"
		ADD CONSTRAINT "sub_shelf_root_shelf_id_fkey"
		FOREIGN KEY (root_shelf_id) REFERENCES "RootShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "MaterialTable"
		ADD CONSTRAINT "material_parent_sub_shelf_id_fkey"
		FOREIGN KEY (parent_sub_shelf_id) REFERENCES "SubShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "BlockPackTable"
		ADD CONSTRAINT "block_pack_parent_sub_shelf_id_fkey"
		FOREIGN KEY (parent_sub_shelf_id) REFERENCES "SubShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "ItemTable"
		ADD CONSTRAINT "item_parent_sub_shelf_id_fkey"
		FOREIGN KEY (parent_sub_shelf_id) REFERENCES "SubShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "ItemTable"
		ADD CONSTRAINT "item_root_shelf_id_fkey"
		FOREIGN KEY (root_shelf_id) REFERENCES "RootShelfTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;
END
$$;
