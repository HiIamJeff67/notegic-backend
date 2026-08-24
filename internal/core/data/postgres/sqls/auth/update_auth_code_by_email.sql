-- name: UpdateAuthCodeForSendingValidationEmail
UPDATE "UserAccountTable" ua
SET
    auth_code = $1, 
    auth_code_expired_at = $2, 
    block_auth_code_until = $3
FROM "UserTable" u
WHERE ua.user_id = u.id AND u.email = $4 AND block_auth_code_until < now()
RETURNING u.name AS name, u.user_agent AS user_agent, block_auth_code_until AS block_auth_code_until, now() AS now;
