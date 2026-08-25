-- name: MoveMyBlockPackById
UPDATE "BlockPackTable" bp
SET 
    "parent_sub_shelf_id" = $1, 
    "updated_at" = NOW()
FROM "SubShelfTable" ss
JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id
WHERE 
    bp.id = $2
    AND bp.deleted_at IS NULL
    AND bp.parent_sub_shelf_id = ss.id
    AND uts.user_id = $3
    AND uts.permission IN ($4::"AccessControlPermission"[])
    AND EXISTS (
        SELECT 1
        FROM "SubShelfTable" dest_ss
        JOIN "UsersToShelvesTable" dest_uts ON dest_ss.root_shelf_id = dest_uts.root_shelf_id
        WHERE 
            dest_ss.id = $5
            AND dest_uts.user_id = $6
            AND dest_uts.permission IN ($7::"AccessControlPermission"[])
    )