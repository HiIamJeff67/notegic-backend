-- name: GetUserDataCacheById
SELECT 
    u.id,
    u.public_id,
    u.name, 
    u.display_name,
    u.email,
    u.role,
    u.plan,
    u.status,
    ui.avatar_url,
    u.created_at,
    NOW() AS updated_at
FROM "UserTable" u
LEFT JOIN "UserInfoTable" ui ON u.id = ui.user_id
LEFT JOIN "UserSettingTable" us ON u.id = us.user_id
WHERE u.id = $1
