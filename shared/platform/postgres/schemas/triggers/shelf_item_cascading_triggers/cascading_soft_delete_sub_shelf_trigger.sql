CREATE OR REPLACE FUNCTION trigger_function_cascading_soft_delete_sub_shelf()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE "MaterialTable"
        SET deleted_at = NEW.deleted_at
        WHERE parent_sub_shelf_id = NEW.id AND deleted_at IS NULL;

        UPDATE "BlockPackTable"
        SET deleted_at = NEW.deleted_at
        WHERE parent_sub_shelf_id = NEW.id AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_cascading_soft_delete_sub_shelf ON "SubShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_cascading_soft_delete_sub_shelf
    AFTER UPDATE 
    ON "SubShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_cascading_soft_delete_sub_shelf();
