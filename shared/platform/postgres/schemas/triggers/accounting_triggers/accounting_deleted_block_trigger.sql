CREATE OR REPLACE FUNCTION trigger_function_accounting_deleted_block()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
    IF (TG_OP <> 'DELETE') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_deleted_block: %. Expected DELETE.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    PERFORM 1
    FROM "RootShelfTable" rs
    JOIN "SubShelfTable" ss ON ss.root_shelf_id = rs.id
    JOIN "BlockPackTable" bp ON bp.parent_sub_shelf_id = ss.id
    JOIN old_table ob ON ob.block_pack_id = bp.id
    ORDER BY rs.id
    FOR UPDATE OF rs;

    WITH old_blocks_agg AS (
        SELECT 
            block_pack_id, 
            count(*) as count_delta
        FROM old_table
        GROUP BY block_pack_id
    ),
    owner_deltas AS (
        SELECT 
            owner_uts.user_id AS owner_id,
            sum(oba.count_delta) as total_delta
        FROM old_blocks_agg oba
        JOIN "BlockPackTable" bp ON oba.block_pack_id = bp.id
        JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
        JOIN "UsersToShelvesTable" owner_uts ON ss.root_shelf_id = owner_uts.root_shelf_id AND owner_uts.permission = 'Owner'
        GROUP BY owner_uts.user_id
    )
    UPDATE "UserAccountTable" ua
    SET 
        block_count = GREATEST(0, block_count - od.total_delta),
        updated_at = NOW()
    FROM owner_deltas od
    WHERE ua.user_id = od.owner_id;

    WITH old_blocks_agg AS (
        SELECT 
            block_pack_id, 
            count(*) as count_delta
        FROM old_table
        GROUP BY block_pack_id
    ),
    bp_updates AS (
        SELECT 
            oba.block_pack_id, 
            sum(oba.count_delta) as total_delta
        FROM old_blocks_agg oba
        GROUP BY oba.block_pack_id
    )
    UPDATE "BlockPackTable" bp
    SET 
        block_count = GREATEST(0, block_count - bpu.total_delta),
        updated_at = NOW()
    FROM bp_updates bpu
    WHERE bp.id = bpu.block_pack_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog, public;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_deleted_block ON "BlockTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_deleted_block
    AFTER DELETE 
    ON "BlockTable"
    REFERENCING OLD TABLE AS old_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_deleted_block();
