CREATE OR REPLACE FUNCTION trigger_function_cascading_restore_soft_deleted_root_shelf()
RETURNS TRIGGER AS $$
DECLARE
    updated_sub_shelf_ids uuid[];
BEGIN
    IF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
        WITH updated_sub_shelves AS (
            UPDATE "SubShelfTable"
            SET deleted_at = NULL
            WHERE root_shelf_id = NEW.id 
            AND deleted_at = OLD.deleted_at
            RETURNING id
        )
        SELECT COALESCE(array_agg(id), ARRAY[]::uuid[])
        INTO updated_sub_shelf_ids
        FROM updated_sub_shelves;

        UPDATE "MaterialTable"
        SET deleted_at = NULL
        WHERE "MaterialTable".parent_sub_shelf_id = ANY(updated_sub_shelf_ids)
        AND "MaterialTable".deleted_at = OLD.deleted_at;

        UPDATE "BlockPackTable"
        SET deleted_at = NULL
        WHERE "BlockPackTable".parent_sub_shelf_id = ANY(updated_sub_shelf_ids)
        AND "BlockPackTable".deleted_at = OLD.deleted_at;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_cascading_restore_soft_deleted_root_shelf ON "RootShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_cascading_restore_soft_deleted_root_shelf
    AFTER UPDATE 
    ON "RootShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_cascading_restore_soft_deleted_root_shelf();
