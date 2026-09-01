CREATE OR REPLACE FUNCTION trigger_function_project_materials_to_items_after_insert_or_update()
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
        material.id,
        material.parent_sub_shelf_id,
        sub_shelf.root_shelf_id,
        'Material'::"ItemType",
        material.deleted_at,
        material.updated_at,
        material.created_at
    FROM new_rows AS material
    JOIN "SubShelfTable" AS sub_shelf
      ON sub_shelf.id = material.parent_sub_shelf_id
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

CREATE OR REPLACE FUNCTION trigger_function_delete_material_items_after_delete()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM "ItemTable" item
    USING old_rows AS material
    WHERE item.id = material.id
      AND item.type = 'Material'::"ItemType";

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_project_materials_to_items_after_insert ON "MaterialTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_project_materials_to_items_after_insert
    AFTER INSERT
    ON "MaterialTable"
    REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_project_materials_to_items_after_insert_or_update();

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_project_materials_to_items_after_update ON "MaterialTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_project_materials_to_items_after_update
    AFTER UPDATE
    ON "MaterialTable"
    REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_project_materials_to_items_after_insert_or_update();

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_delete_material_items_after_delete ON "MaterialTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_delete_material_items_after_delete
    AFTER DELETE
    ON "MaterialTable"
    REFERENCING OLD TABLE AS old_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_delete_material_items_after_delete();
