CREATE OR REPLACE FUNCTION trigger_function_accounting_mutated_material()
RETURNS TRIGGER AS $$
DECLARE
    current_count BIGINT;
    max_count INTEGER;
    current_count_per_root_shelf BIGINT;
    max_count_per_root_shelf INTEGER;
    plan_name TEXT;
    root_shelf_id UUID;
    owner_id UUID;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        SELECT
            pl.max_material_count,
            pl.max_item_count_per_root_shelf,
            u.plan::TEXT, 
            rs.id,
            u.id
        INTO
            max_count,
            max_count_per_root_shelf,
            plan_name,
            root_shelf_id,
            owner_id
        FROM "SubShelfTable" ss
        JOIN "RootShelfTable" rs ON ss.root_shelf_id = rs.id
        JOIN "UserTable" u ON rs.owner_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE ss.id = NEW.parent_sub_shelf_id
        FOR UPDATE OF rs;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find owner for Material (SubShelf ID: %). Possible orphan record.', NEW.parent_sub_shelf_id
            USING ERRCODE = 'data_exception';
        END IF;

        UPDATE "UserAccountTable"
        SET
            material_count = material_count + 1,
            updated_at = NOW()
        WHERE user_id = owner_id
        RETURNING material_count INTO current_count;

        IF current_count > max_count THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % materials. Current count: %.', 
                plan_name, max_count, current_count
            USING ERRCODE = 'check_violation';
        END IF;

        UPDATE "RootShelfTable"
        SET
            item_count = item_count + 1,
            updated_at = NOW()
        WHERE id = root_shelf_id
        RETURNING item_count INTO current_count_per_root_shelf;

        IF current_count_per_root_shelf > max_count_per_root_shelf THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % items per root shelf. Current count: %.', 
                plan_name, max_count_per_root_shelf, current_count_per_root_shelf
            USING ERRCODE = 'check_violation';
        END IF;

        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        PERFORM 1
        FROM "SubShelfTable" ss
        JOIN "RootShelfTable" rs ON ss.root_shelf_id = rs.id
        WHERE ss.id = OLD.parent_sub_shelf_id
        FOR UPDATE OF rs;

        UPDATE "UserAccountTable"
        SET
            material_count = GREATEST(0, material_count - 1),
            updated_at = NOW()
        FROM "SubShelfTable" ss, "RootShelfTable" rs
        WHERE ss.id = OLD.parent_sub_shelf_id
        AND ss.root_shelf_id = rs.id
        AND user_id = rs.owner_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find UserAccount of the owner for Material (SubShelf ID: %). Possible orphan record.', OLD.parent_sub_shelf_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        UPDATE "RootShelfTable"
        SET
            item_count = GREATEST(0, item_count - 1),
            updated_at = NOW()
        FROM "SubShelfTable" ss
        WHERE ss.id = OLD.parent_sub_shelf_id
        AND "RootShelfTable".id = ss.root_shelf_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find RootShelf for Material (SubShelf ID: %). Possible orphan record.', OLD.parent_sub_shelf_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_mutated_material ON "MaterialTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_mutated_material
    BEFORE INSERT OR DELETE 
    ON "MaterialTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_mutated_material();
