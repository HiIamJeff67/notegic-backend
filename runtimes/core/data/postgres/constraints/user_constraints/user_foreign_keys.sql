DO $$
DECLARE
	constraint_record RECORD;
BEGIN
	FOR constraint_record IN
		SELECT tables.relname AS table_name, constraints.conname
		FROM pg_constraint constraints
		JOIN pg_class tables ON tables.oid = constraints.conrelid
		JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
		WHERE schemas.nspname = 'public'
		  AND constraints.contype = 'f'
		  AND (
			(tables.relname = 'UserInfoTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'UserAccountTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'UserSettingTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'APIKeyTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'ThemeTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (author_id)%')
			OR (tables.relname = 'UsersToBadgesTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'UsersToBadgesTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (badge_id)%')
			OR (tables.relname = 'UsersToShelvesTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'UsersToStationsTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'RoutineTagTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (owner_id)%')
			OR (tables.relname = 'StationTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (owner_id)%')
			OR (tables.relname = 'UsersToBillingPlansTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (user_id)%')
			OR (tables.relname = 'UsersToBillingPlansTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (billing_plan_id)%')
			OR (tables.relname = 'RootShelfTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (owner_id)%')
			OR (tables.relname = 'UserTable' AND pg_get_constraintdef(constraints.oid) LIKE 'FOREIGN KEY (plan)%')
		  )
	LOOP
		EXECUTE format(
			'ALTER TABLE %I DROP CONSTRAINT %I',
			constraint_record.table_name,
			constraint_record.conname
		);
	END LOOP;

	ALTER TABLE "UserInfoTable"
		ADD CONSTRAINT "user_info_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UserAccountTable"
		ADD CONSTRAINT "user_account_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UserSettingTable"
		ADD CONSTRAINT "user_setting_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "APIKeyTable"
		ADD CONSTRAINT "api_key_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "ThemeTable"
		ADD CONSTRAINT "theme_author_id_fkey"
		FOREIGN KEY (author_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToBadgesTable"
		ADD CONSTRAINT "users_to_badges_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToBadgesTable"
		ADD CONSTRAINT "users_to_badges_badge_id_fkey"
		FOREIGN KEY (badge_id) REFERENCES "BadgeTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToShelvesTable"
		ADD CONSTRAINT "users_to_shelves_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToStationsTable"
		ADD CONSTRAINT "users_to_stations_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RoutineTagTable"
		ADD CONSTRAINT "routine_tag_owner_id_fkey"
		FOREIGN KEY (owner_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "StationTable"
		ADD CONSTRAINT "station_owner_id_fkey"
		FOREIGN KEY (owner_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToBillingPlansTable"
		ADD CONSTRAINT "users_to_billing_plans_user_id_fkey"
		FOREIGN KEY (user_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UsersToBillingPlansTable"
		ADD CONSTRAINT "users_to_billing_plans_billing_plan_id_fkey"
		FOREIGN KEY (billing_plan_id) REFERENCES "BillingPlanTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "RootShelfTable"
		ADD CONSTRAINT "root_shelf_owner_id_fkey"
		FOREIGN KEY (owner_id) REFERENCES "UserTable" (id)
		ON UPDATE CASCADE ON DELETE CASCADE;

	ALTER TABLE "UserTable"
		ADD CONSTRAINT "user_plan_fkey"
		FOREIGN KEY (plan) REFERENCES "PlanLimitationTable" (key)
		ON UPDATE CASCADE ON DELETE CASCADE;
END
$$;
