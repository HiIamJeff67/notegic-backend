CREATE OR REPLACE FUNCTION trigger_function_project_sub_shelves_to_items()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE "ItemTable" item
    SET root_shelf_id = new_sub_shelves.root_shelf_id
    FROM new_rows AS new_sub_shelves
    JOIN old_rows AS old_sub_shelves
      ON old_sub_shelves.id = new_sub_shelves.id
    WHERE item.parent_sub_shelf_id = new_sub_shelves.id
      AND old_sub_shelves.root_shelf_id IS DISTINCT FROM new_sub_shelves.root_shelf_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_project_sub_shelves_to_items ON "SubShelfTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_project_sub_shelves_to_items
    AFTER UPDATE
    ON "SubShelfTable"
    REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_project_sub_shelves_to_items();
