CREATE OR REPLACE FUNCTION trigger_function_accounting_mutated_sub_shelf()
RETURNS TRIGGER AS $$
DECLARE
    current_count BIGINT;
    max_count INTEGER;
    plan_name TEXT;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        SELECT
            pl.max_sub_shelf_count_per_root_shelf, 
            u.plan::TEXT
        INTO
            max_count,
            plan_name
        FROM "RootShelfTable" rs
        JOIN "UserTable" u ON rs.owner_id = u.id
        JOIN "PlanLimitationTable" pl ON u.plan = pl.key
        WHERE rs.id = NEW.root_shelf_id
        FOR UPDATE OF rs;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find owner for SubShelf (RootShelf ID: %). Possible orphan record.', NEW.root_shelf_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        UPDATE "RootShelfTable"
        SET
            sub_shelf_count = sub_shelf_count + 1,
            updated_at = NOW()
        WHERE id = NEW.root_shelf_id
        RETURNING sub_shelf_count INTO current_count;

        IF current_count > max_count THEN
            RAISE EXCEPTION 'Quota exceeded: Plan "%" allows maximum % sub shelves per root shelf. Current count: %.', 
                plan_name, max_count, current_count
            USING ERRCODE = 'check_violation';
        END IF;

        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE "RootShelfTable"
        SET
            sub_shelf_count = GREATEST(0, sub_shelf_count - 1),
            updated_at = NOW()
        WHERE id = OLD.root_shelf_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Data integrity: Cannot find RootShelf for SubShelf (RootShelf ID: %). Possible orphan record.', OLD.root_shelf_id
            USING ERRCODE = 'integrity_constraint_violation';
        END IF;

        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_accounting_mutated_sub_shelf ON "SubShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_accounting_mutated_sub_shelf
    BEFORE INSERT OR DELETE 
    ON "SubShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_accounting_mutated_sub_shelf();
