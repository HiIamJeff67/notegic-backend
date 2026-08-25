ALTER TABLE "UserQuotaTable" DROP CONSTRAINT IF EXISTS user_quota_table_pkey;
DROP INDEX IF EXISTS user_quota_table_user_id_unique;

-- ============================== SQL Separator ==============================

ALTER TABLE "UserQuotaTable" ADD COLUMN IF NOT EXISTS "id" uuid;

-- ============================== SQL Separator ==============================

UPDATE "UserQuotaTable" SET "id" = gen_random_uuid() WHERE "id" IS NULL;

-- ============================== SQL Separator ==============================

ALTER TABLE "UserQuotaTable" ALTER COLUMN "id" SET DEFAULT gen_random_uuid();

-- ============================== SQL Separator ==============================

ALTER TABLE "UserQuotaTable" ALTER COLUMN "id" SET NOT NULL;

-- ============================== SQL Separator ==============================

ALTER TABLE "UserQuotaTable" ADD CONSTRAINT user_quota_table_pkey PRIMARY KEY ("id");
CREATE UNIQUE INDEX IF NOT EXISTS user_quota_table_user_id_unique ON "UserQuotaTable" ("user_id");
