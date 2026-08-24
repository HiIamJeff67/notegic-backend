CREATE OR REPLACE FUNCTION trigger_function_project_block_packs_to_items_after_insert_or_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO "ItemTable" (
        id,
        parent_sub_shelf_id,
        root_shelf_id,
        type,
        deleted_at,
        updated_at,
        created_at
    )
    SELECT
        block_pack.id,
        block_pack.parent_sub_shelf_id,
        sub_shelf.root_shelf_id,
        'BlockPack'::"ItemType",
        block_pack.deleted_at,
        block_pack.updated_at,
        block_pack.created_at
    FROM new_rows AS block_pack
    JOIN "SubShelfTable" AS sub_shelf
      ON sub_shelf.id = block_pack.parent_sub_shelf_id
    ON CONFLICT (id, type) DO UPDATE SET
        parent_sub_shelf_id = EXCLUDED.parent_sub_shelf_id,
        root_shelf_id = EXCLUDED.root_shelf_id,
        deleted_at = EXCLUDED.deleted_at,
        updated_at = EXCLUDED.updated_at,
        created_at = EXCLUDED.created_at;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

CREATE OR REPLACE FUNCTION trigger_function_delete_block_pack_items_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM "ItemTable" item
    USING old_rows AS block_pack
    WHERE item.id = block_pack.id
      AND item.type = 'BlockPack'::"ItemType";

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_project_block_packs_to_items_after_insert ON "BlockPackTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_project_block_packs_to_items_after_insert
    AFTER INSERT
    ON "BlockPackTable"
    REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_project_block_packs_to_items_after_insert_or_update();

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_project_block_packs_to_items_after_update ON "BlockPackTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_project_block_packs_to_items_after_update
    AFTER UPDATE
    ON "BlockPackTable"
    REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_project_block_packs_to_items_after_insert_or_update();

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_delete_block_pack_items_after_delete ON "BlockPackTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_delete_block_pack_items_after_delete
    AFTER DELETE
    ON "BlockPackTable"
    REFERENCING OLD TABLE AS old_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_delete_block_pack_items_after_delete();
