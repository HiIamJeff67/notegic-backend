CREATE OR REPLACE FUNCTION trigger_function_accounting_inserted_block()
RETURNS TRIGGER AS $$
DECLARE
    r RECORD;
BEGIN
    IF (TG_OP <> 'INSERT') THEN
        RAISE EXCEPTION 'Invalid operation for trigger_function_accounting_inserted_block: %. Expected INSERT.', TG_OP
        USING ERRCODE = 'program_limit_exceeded';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM new_table nb
        LEFT JOIN "BlockPackTable" bp ON bp.id = nb.block_pack_id
        LEFT JOIN "SubShelfTable" ss ON ss.id = bp.parent_sub_shelf_id
        LEFT JOIN "RootShelfTable" rs ON rs.id = ss.root_shelf_id
        LEFT JOIN "UsersToShelvesTable" owner_uts
            ON owner_uts.root_shelf_id = rs.id
            AND owner_uts.permission = 'Owner'
        LEFT JOIN "UserTable" u ON u.id = owner_uts.user_id
        LEFT JOIN "PlanLimitationTable" pl ON pl.key = u.plan
        LEFT JOIN "UserAccountTable" ua ON ua.user_id = owner_uts.user_id
        WHERE bp.id IS NULL
            OR ss.id IS NULL
            OR rs.id IS NULL
            OR owner_uts.user_id IS NULL
            OR u.id IS NULL
            OR pl.key IS NULL
            OR ua.user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'Data integrity: Cannot find owner or plan limitation for Block.'
        USING ERRCODE = 'data_exception';
    END IF;

    PERFORM 1
    FROM "RootShelfTable" rs
    JOIN "SubShelfTable" ss ON ss.root_shelf_id = rs.id
    JOIN "BlockPackTable" bp ON bp.parent_sub_shelf_id = ss.id
    JOIN new_table nb ON nb.block_pack_id = bp.id
    ORDER BY rs.id
    FOR UPDATE OF rs;

    FOR r IN -- batch account the user block count, if all the mutated blocks belong to one user, then this will only execute once
        WITH new_blocks_agg AS (
            SELECT 
                block_pack_id, 
                count(*) as count_delta
            FROM new_table
            GROUP BY block_pack_id
        ),
        owner_deltas AS (
            SELECT 
                owner_uts.user_id AS owner_id,
                sum(nba.count_delta) as total_delta
            FROM new_blocks_agg nba
            JOIN "BlockPackTable" bp ON nba.block_pack_id = bp.id
            JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
            JOIN "UsersToShelvesTable" owner_uts ON ss.root_shelf_id = owner_uts.root_shelf_id AND owner_uts.permission = 'Owner'
            GROUP BY owner_uts.user_id
        ),
        updated_accounts AS (
            UPDATE "UserAccountTable" ua
            SET 
                block_count = block_count + od.total_delta,
                updated_at = NOW()
            FROM owner_deltas od
            WHERE ua.user_id = od.owner_id
            RETURNING ua.user_id, ua.block_count
        )

        -- finally select the limitation and iteratively check the updated block count of the user is not exceeded the limitation
        SELECT 
            u.id, u.plan, ua.block_count, pl.max_block_count
        FROM updated_accounts ua
        JOIN "UserTable" u ON ua.user_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE ua.block_count > pl.max_block_count
    LOOP
        RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % blocks. Current count: %.', 
            r.plan, r.max_block_count, r.block_count
        USING ERRCODE = 'check_violation';
    END LOOP;

    FOR r IN
        WITH new_blocks_agg AS (
            SELECT 
                block_pack_id, 
                count(*) as count_delta
            FROM new_table
            GROUP BY block_pack_id
        ),
        bp_deltas AS (
            SELECT 
                nba.block_pack_id, 
                sum(nba.count_delta) as total_delta
            FROM new_blocks_agg nba
            GROUP BY nba.block_pack_id
        ),
        updated_packs AS (
            UPDATE "BlockPackTable" bp
            SET 
                block_count = block_count + bpu.total_delta,
                updated_at = NOW()
            FROM bp_deltas bpu
            WHERE bp.id = bpu.block_pack_id
            RETURNING bp.id, bp.block_count
        )

        SELECT DISTINCT ON (up.id)
            up.id, up.block_count, pl.max_block_count_per_block_pack, u.plan
        FROM updated_packs up
        JOIN "BlockPackTable" bp ON bp.id = up.id
        JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
        JOIN "UsersToShelvesTable" owner_uts ON ss.root_shelf_id = owner_uts.root_shelf_id AND owner_uts.permission = 'Owner'
        JOIN "UserTable" u ON owner_uts.user_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE up.block_count > pl.max_block_count_per_block_pack
    LOOP
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % blocks in each block pack. Current count: %.', 
            r.plan, r.max_block_count_per_block_pack, r.block_count
        USING ERRCODE = 'check_violation';
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog, public;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_inserted_block ON "BlockTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_inserted_block
    AFTER INSERT 
    ON "BlockTable"
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT
    EXECUTE FUNCTION trigger_function_accounting_inserted_block();
