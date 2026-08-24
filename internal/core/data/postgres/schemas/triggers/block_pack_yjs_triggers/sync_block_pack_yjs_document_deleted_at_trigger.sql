CREATE OR REPLACE FUNCTION trigger_function_sync_block_pack_yjs_document_deleted_at()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS DISTINCT FROM NEW.deleted_at THEN
        UPDATE "BlockPackYjsDocumentTable"
        SET
            deleted_at = NEW.deleted_at,
            updated_at = NOW()
        WHERE block_pack_id = NEW.id
        AND deleted_at IS DISTINCT FROM NEW.deleted_at;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_sync_block_pack_yjs_document_deleted_at ON "BlockPackTable";

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_sync_block_pack_yjs_document_deleted_at
    AFTER UPDATE OF deleted_at
    ON "BlockPackTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_sync_block_pack_yjs_document_deleted_at();
