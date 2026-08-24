-- name: DeleteAllMyBadges
DELETE FROM "UsersToBadgesTable" utb
WHERE utb.user_id IN ($1)
RETURNING NOW() AS deleted_at;