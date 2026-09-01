CREATE OR REPLACE VIEW "UserView" AS
SELECT
    "id",
    "public_id",
    "plan",
    "status",
    "created_at"
FROM "UserTable";
