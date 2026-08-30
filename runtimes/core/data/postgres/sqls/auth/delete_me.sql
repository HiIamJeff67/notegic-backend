-- name: DeleteMe
DELETE FROM "UserTable" u
USING "UserAccountTable" ua
WHERE u.id = $1
    AND ua.user_id = u.id
    AND (
        u.role = 'Guest'
        OR (
            ua.auth_code = $2
            AND ua.auth_code_expired_at > NOW()
            AND u.role <> 'Guest'
        )
    )
RETURNING NOW() AS deleted_at;