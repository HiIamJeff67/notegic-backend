-- name: ResetEmail
UPDATE "UserTable" u
SET 
    email = $1
FROM "UserAccountTable" ua
WHERE ua.auth_code = $2
    AND u.user_id = $3
    AND u.user_id = ua.user_id
RETURNING u.updated_at;