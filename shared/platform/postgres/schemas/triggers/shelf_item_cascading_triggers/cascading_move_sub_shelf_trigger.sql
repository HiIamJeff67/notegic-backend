CREATE OR REPLACE FUNCTION trigger_function_cascading_move_sub_shelf()
RETURNS TRIGGER AS $$
DECLARE
    child RECORD;
    source_index INTEGER;
    child_path_suffix uuid[];
    new_child_path uuid[];

    child_ids uuid[] := ARRAY[]::uuid[];
    child_new_paths TEXT[] := ARRAY[]::TEXT[];
BEGIN
    IF OLD.prev_sub_shelf_id IS DISTINCT FROM NEW.prev_sub_shelf_id
    OR OLD.root_shelf_id IS DISTINCT FROM NEW.root_shelf_id THEN

        FOR child IN
            SELECT id, path
            FROM "SubShelfTable"
            WHERE root_shelf_id = OLD.root_shelf_id
            AND path @> ARRAY[NEW.id]::uuid[]
            AND id != NEW.id
            AND deleted_at IS NULL
        LOOP
            source_index := NULL;
            FOR i IN 1..array_length(child.path, 1) LOOP
                IF child.path[i] = NEW.id THEN
                    source_index := i;
                    EXIT;
                END IF;
            END LOOP;

            IF source_index IS NOT NULL THEN
                child_path_suffix := child.path[source_index:array_length(child.path, 1)];

                new_child_path := NEW.path || child_path_suffix;

                child_ids := child_ids || child.id;
                child_new_paths := child_new_paths || ('{' || array_to_string(new_child_path, ',') || '}');
            END IF;
        END LOOP;

        IF array_length(child_ids, 1) > 0 THEN
            UPDATE "SubShelfTable"
            SET
                root_shelf_id = NEW.root_shelf_id,
                path = child_sub_shelf.new_path::uuid[], 
                updated_at = NOW()
            FROM (
                SELECT
                    unnest(child_ids) AS id, 
                    unnest(child_new_paths) AS new_path
            ) AS child_sub_shelf
            WHERE "SubShelfTable".id = child_sub_shelf.id
                AND deleted_at IS NULL;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================== SQL Separator ==============================

DROP TRIGGER IF EXISTS trigger_cascading_move_sub_shelf ON "SubShelfTable"

-- ============================== SQL Separator ==============================

CREATE TRIGGER trigger_cascading_move_sub_shelf
    AFTER UPDATE OF prev_sub_shelf_id, root_shelf_id
    ON "SubShelfTable"
    FOR EACH ROW
    EXECUTE FUNCTION trigger_function_cascading_move_sub_shelf();
