-- name: ValidateEmailByAuthCode
UPDATE "UserTable" u
SET role = 'Normal'
FROM "UserAccountTable" ua
WHERE ua.user_id = u.id
    AND u.id = $1
    AND ua.auth_code = $2
    AND ua.auth_code_expired_at > NOW()
RETURNING u.updated_at AS updated_at;