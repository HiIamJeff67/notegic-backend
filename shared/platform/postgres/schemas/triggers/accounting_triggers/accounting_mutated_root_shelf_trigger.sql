CREATE OR REPLACE FUNCTION trigger_function_accounting_mutated_root_shelf()
RETURNS TRIGGER AS $$
DECLARE
    current_count BIGINT;
    max_count INTEGER;
    max_root_shelf_count INTEGER;
    max_block_pack_count INTEGER;
    max_material_count INTEGER;
    max_block_count INTEGER;
    plan_name TEXT;
    block_pack_delta BIGINT;
    material_delta BIGINT;
    block_delta BIGINT;
    max_sub_shelf_count INTEGER;
    max_item_count INTEGER;
    max_block_count_per_block_pack INTEGER;
    largest_block_pack_count BIGINT;
    transferred_root_shelf_count BIGINT;
    transferred_block_pack_count BIGINT;
    transferred_material_count BIGINT;
    transferred_block_count BIGINT;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        SELECT
            pl.max_root_shelf_count,
            u.plan::TEXT
        INTO
            max_count,
            plan_name
        FROM "UserTable" u
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE u.id = NEW.owner_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find owner for RootShelf. Possible orphan record.'
            USING ERRCODE = 'data_exception';
        END IF;

        UPDATE "UserAccountTable" ua
        SET 
            root_shelf_count = root_shelf_count + 1,
            updated_at = NOW()
        WHERE ua.user_id = NEW.owner_id
        RETURNING root_shelf_count INTO current_count;

        IF current_count > max_count THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % root shelves. Current count: %.', 
                plan_name, max_count, current_count
            USING ERRCODE = 'check_violation';
        END IF;

        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE "UserAccountTable"
        SET
            root_shelf_count = GREATEST(0, root_shelf_count - 1),
            updated_at = NOW()
        WHERE user_id = OLD.owner_id;

        IF NOT FOUND THEN
             RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the owner for RootShelf (Owner ID: %).', OLD.owner_id
             USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        RETURN OLD;

    ELSIF (TG_OP = 'UPDATE') THEN
        IF NEW.owner_id IS NOT DISTINCT FROM OLD.owner_id THEN
            RETURN NEW;
        END IF;

        SELECT
            pl.max_root_shelf_count,
            pl.max_block_pack_count,
            pl.max_material_count,
            pl.max_block_count,
            pl.max_sub_shelf_count_per_root_shelf,
            pl.max_item_count_per_root_shelf,
            pl.max_block_count_per_block_pack,
            u.plan::TEXT
        INTO
            max_root_shelf_count,
            max_block_pack_count,
            max_material_count,
            max_block_count,
            max_sub_shelf_count,
            max_item_count,
            max_block_count_per_block_pack,
            plan_name
        FROM "UserTable" u
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE u.id = NEW.owner_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find new owner for RootShelf. Possible orphan record.'
            USING ERRCODE = 'data_exception';
        END IF;

        SELECT count(*) INTO block_pack_delta
        FROM "BlockPackTable" bp
        JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
        WHERE ss.root_shelf_id = NEW.id;

        SELECT count(*) INTO material_delta
        FROM "MaterialTable" m
        JOIN "SubShelfTable" ss ON m.parent_sub_shelf_id = ss.id
        WHERE ss.root_shelf_id = NEW.id;

        SELECT count(*) INTO block_delta
        FROM "BlockTable" b
        JOIN "BlockPackTable" bp ON b.block_pack_id = bp.id
        JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
        WHERE ss.root_shelf_id = NEW.id;

        SELECT COALESCE(max(bp.block_count), 0) INTO largest_block_pack_count
        FROM "BlockPackTable" bp
        JOIN "SubShelfTable" ss ON bp.parent_sub_shelf_id = ss.id
        WHERE ss.root_shelf_id = NEW.id;

        IF NEW.sub_shelf_count > max_sub_shelf_count THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % sub shelves per root shelf. Current count: %.',
                plan_name, max_sub_shelf_count, NEW.sub_shelf_count
            USING ERRCODE = 'check_violation';
        END IF;
        IF NEW.item_count > max_item_count THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % items per root shelf. Current count: %.',
                plan_name, max_item_count, NEW.item_count
            USING ERRCODE = 'check_violation';
        END IF;
        IF largest_block_pack_count > max_block_count_per_block_pack THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % blocks in each block pack. Current count: %.',
                plan_name, max_block_count_per_block_pack, largest_block_pack_count
            USING ERRCODE = 'check_violation';
        END IF;

        UPDATE "UserAccountTable"
        SET
            root_shelf_count = GREATEST(0, root_shelf_count - 1),
            block_pack_count = GREATEST(0, block_pack_count - block_pack_delta),
            material_count = GREATEST(0, material_count - material_delta),
            block_count = GREATEST(0, block_count - block_delta),
            updated_at = NOW()
        WHERE user_id = OLD.owner_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the previous RootShelf owner (Owner ID: %).', OLD.owner_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        UPDATE "UserAccountTable"
        SET
            root_shelf_count = root_shelf_count + 1,
            block_pack_count = block_pack_count + block_pack_delta,
            material_count = material_count + material_delta,
            block_count = block_count + block_delta,
            updated_at = NOW()
        WHERE user_id = NEW.owner_id
        RETURNING root_shelf_count, block_pack_count, material_count, block_count
        INTO transferred_root_shelf_count, transferred_block_pack_count, transferred_material_count, transferred_block_count;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the new RootShelf owner (Owner ID: %).', NEW.owner_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        IF transferred_root_shelf_count > max_root_shelf_count
            OR transferred_block_pack_count > max_block_pack_count
            OR transferred_material_count > max_material_count
            OR transferred_block_count > max_block_count THEN
            RAISE EXCEPTION 'Quota exceeded while transferring RootShelf ownership to plan "%".', plan_name
            USING ERRCODE = 'check_violation';
        END IF;

        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_mutated_root_shelf ON "RootShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_mutated_root_shelf
    BEFORE INSERT OR DELETE OR UPDATE
    ON "RootShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_mutated_root_shelf();
