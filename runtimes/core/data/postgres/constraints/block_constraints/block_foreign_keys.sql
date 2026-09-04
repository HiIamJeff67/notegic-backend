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
		  AND tables.relname IN (
			'BlockPackYjsDocumentTable',
			'BlockPackYjsUpdateTable',
			'BlockTable'
		  )
		  AND constraints.contype = 'f'
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			constraint_record.table_name,
			constraint_record.conname
		);
	END LOOP;

	ALTER TABLE "BlockPackYjsDocumentTable"
		ADD CONSTRAINT "block_pack_yjs_document_block_pack_id_fkey"
		FOREIGN KEY (block_pack_id) REFERENCES "BlockPackTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "BlockPackYjsUpdateTable"
		ADD CONSTRAINT "block_pack_yjs_update_block_pack_id_fkey"
		FOREIGN KEY (block_pack_id) REFERENCES "BlockPackTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "BlockTable"
		ADD CONSTRAINT "block_block_pack_id_fkey"
		FOREIGN KEY (block_pack_id) REFERENCES "BlockPackTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "BlockTable"
		ADD CONSTRAINT "block_parent_block_id_fkey"
		FOREIGN KEY (parent_block_id) REFERENCES "BlockTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "BlockTable"
		ADD CONSTRAINT "block_prev_block_id_fkey"
		FOREIGN KEY (prev_block_id) REFERENCES "BlockTable" (id)
		ON UPDATE CASCADE ON DELETE SET NULL;

	ALTER TABLE "BlockTable"
		ADD CONSTRAINT "block_next_block_id_fkey"
		FOREIGN KEY (next_block_id) REFERENCES "BlockTable" (id)
		ON UPDATE CASCADE ON DELETE SET NULL;
END
$$;
