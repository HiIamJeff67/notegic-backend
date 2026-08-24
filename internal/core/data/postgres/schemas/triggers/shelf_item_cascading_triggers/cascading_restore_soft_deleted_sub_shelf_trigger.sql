CREATE OR REPLACE FUNCTION trigger_function_cascading_restore_soft_deleted_sub_shelf()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
        UPDATE "MaterialTable"
        SET deleted_at = NULL
        WHERE parent_sub_shelf_id = NEW.id
        AND deleted_at = OLD.deleted_at;

        UPDATE "BlockPackTable"
        SET deleted_at = NULL
        WHERE parent_sub_shelf_id = NEW.id
        AND deleted_at = OLD.deleted_at;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_cascading_restore_soft_deleted_sub_shelf ON "SubShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_cascading_restore_soft_deleted_sub_shelf
    AFTER UPDATE 
    ON "SubShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_cascading_restore_soft_deleted_sub_shelf();
