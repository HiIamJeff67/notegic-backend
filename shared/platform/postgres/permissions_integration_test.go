package postgres

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

type migrationPermissionProbe struct {
	Id string `gorm:"column:id;primaryKey"`
}

func (migrationPermissionProbe) TableName() string {
	return "permission_migration_probe"
}

func TestRuntimePermissionsIntegration(t *testing.T) {
	if os.Getenv("NOTEGIC_RUN_POSTGRES_PERMISSION_INTEGRATION") != "1" {
		t.Skip("set NOTEGIC_RUN_POSTGRES_PERMISSION_INTEGRATION=1 to run PostgreSQL permission integration tests")
	}

	admin := integrationPostgresConfig(t, "POSTGRES_PERMISSION_ADMIN")
	core := integrationPostgresConfig(t, "POSTGRES_PERMISSION_CORE")
	durableJob := integrationPostgresConfig(t, "POSTGRES_PERMISSION_DURABLEJOB")
	adminDB, err := Connect(admin)
	if err != nil {
		t.Fatalf("connect admin database: %v", err)
	}
	defer Disconnect(adminDB)

	if err := EnsureRuntimeRole(adminDB, ctypes.Runtime_Core, core.Password); err != nil {
		t.Fatalf("ensure Core role: %v", err)
	}
	if err := EnsureRuntimeRole(adminDB, ctypes.Runtime_DurableJob, durableJob.Password); err != nil {
		t.Fatalf("ensure DurableJob role: %v", err)
	}
	const parentRoleName = "permission_integration_parent"
	const memberRoleName = "permission_integration_member"
	if err := adminDB.Exec("DROP ROLE IF EXISTS " + quoteIdentifier(memberRoleName)).Error; err != nil {
		t.Fatalf("remove membership probe role: %v", err)
	}
	if err := adminDB.Exec("DROP ROLE IF EXISTS " + quoteIdentifier(parentRoleName)).Error; err != nil {
		t.Fatalf("remove parent probe role: %v", err)
	}
	defer adminDB.Exec("DROP ROLE IF EXISTS " + quoteIdentifier(memberRoleName))
	defer adminDB.Exec("DROP ROLE IF EXISTS " + quoteIdentifier(parentRoleName))
	if err := adminDB.Exec("CREATE ROLE " + quoteIdentifier(parentRoleName) + " NOLOGIN").Error; err != nil {
		t.Fatalf("create parent probe role: %v", err)
	}
	if err := adminDB.Exec("CREATE ROLE " + quoteIdentifier(memberRoleName) + " NOLOGIN").Error; err != nil {
		t.Fatalf("create membership probe role: %v", err)
	}
	if err := adminDB.Exec("GRANT " + quoteIdentifier(parentRoleName) + " TO " + quoteIdentifier(ctypes.Runtime_Core.RoleName())).Error; err != nil {
		t.Fatalf("grant parent probe role to Core: %v", err)
	}
	if err := adminDB.Exec("GRANT " + quoteIdentifier(ctypes.Runtime_Core.RoleName()) + " TO " + quoteIdentifier(memberRoleName)).Error; err != nil {
		t.Fatalf("grant Core role to membership probe role: %v", err)
	}
	if err := EnsureRuntimeRole(adminDB, ctypes.Runtime_Core, core.Password); err != nil {
		t.Fatalf("re-harden Core role: %v", err)
	}

	var coreRole struct {
		Inherits   bool `gorm:"column:rolinherit"`
		Superuser  bool `gorm:"column:rolsuper"`
		CreateDB   bool `gorm:"column:rolcreatedb"`
		CreateRole bool `gorm:"column:rolcreaterole"`
		Replicate  bool `gorm:"column:rolreplication"`
		BypassRLS  bool `gorm:"column:rolbypassrls"`
	}
	if err := adminDB.Raw(`
		SELECT rolinherit, rolsuper, rolcreatedb, rolcreaterole, rolreplication, rolbypassrls
		FROM pg_roles
		WHERE rolname = ?
	`, ctypes.Runtime_Core.RoleName()).Scan(&coreRole).Error; err != nil {
		t.Fatalf("read Core role inheritance: %v", err)
	}
	if coreRole.Inherits || coreRole.Superuser || coreRole.CreateDB || coreRole.CreateRole || coreRole.Replicate || coreRole.BypassRLS {
		t.Fatalf("Core role must be least-privileged: %+v", coreRole)
	}

	var runtimeMembershipCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_auth_members memberships
		JOIN pg_roles parent ON parent.oid = memberships.roleid
		JOIN pg_roles member ON member.oid = memberships.member
		WHERE parent.rolname = ? OR member.rolname = ?
	`, ctypes.Runtime_Core.RoleName(), ctypes.Runtime_Core.RoleName()).Scan(&runtimeMembershipCount).Error; err != nil {
		t.Fatalf("read Core role memberships: %v", err)
	}
	if runtimeMembershipCount != 0 {
		t.Fatalf("Core role memberships = %d, want 0", runtimeMembershipCount)
	}

	var publicDatabasePrivilegeCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_database databases
		CROSS JOIN LATERAL aclexplode(COALESCE(databases.datacl, acldefault('d', databases.datdba))) privileges
		WHERE databases.datname = current_database() AND privileges.grantee = 0
	`).Scan(&publicDatabasePrivilegeCount).Error; err != nil {
		t.Fatalf("read PUBLIC database access: %v", err)
	}
	if publicDatabasePrivilegeCount != 0 {
		t.Fatalf("PUBLIC database privileges = %d, want 0", publicDatabasePrivilegeCount)
	}

	var publicSchemaPrivilegeCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_namespace schemas
		CROSS JOIN LATERAL aclexplode(COALESCE(schemas.nspacl, acldefault('n', schemas.nspowner))) privileges
		WHERE schemas.nspname = 'public' AND privileges.grantee = 0
	`).Scan(&publicSchemaPrivilegeCount).Error; err != nil {
		t.Fatalf("read PUBLIC schema access: %v", err)
	}
	if publicSchemaPrivilegeCount != 0 {
		t.Fatalf("PUBLIC schema privileges = %d, want 0", publicSchemaPrivilegeCount)
	}

	var coreCanConnect bool
	if err := adminDB.Raw("SELECT has_database_privilege(?, current_database(), 'CONNECT')", ctypes.Runtime_Core.RoleName()).Scan(&coreCanConnect).Error; err != nil {
		t.Fatalf("read Core database access: %v", err)
	}
	if !coreCanConnect {
		t.Fatal("Core must have database CONNECT")
	}

	var coreCanUseSchema bool
	if err := adminDB.Raw("SELECT has_schema_privilege(?, 'public', 'USAGE')", ctypes.Runtime_Core.RoleName()).Scan(&coreCanUseSchema).Error; err != nil {
		t.Fatalf("read Core schema access: %v", err)
	}
	if !coreCanUseSchema {
		t.Fatal("Core must have schema USAGE")
	}

	var coreCanCreateSchema bool
	if err := adminDB.Raw("SELECT has_schema_privilege(?, 'public', 'CREATE')", ctypes.Runtime_Core.RoleName()).Scan(&coreCanCreateSchema).Error; err != nil {
		t.Fatalf("read Core schema create access: %v", err)
	}
	if coreCanCreateSchema {
		t.Fatal("Core must not retain schema CREATE after role hardening")
	}
	previousLogger := logs.NotegicLogger
	logs.NotegicLogger = logs.NewCommandLineInterfaceLogger()
	defer func() {
		logs.NotegicLogger = previousLogger
	}()
	if err := adminDB.Exec("DROP TABLE IF EXISTS permission_migration_probe").Error; err != nil {
		t.Fatalf("clean up migration probe table: %v", err)
	}
	defer adminDB.Exec("DROP TABLE IF EXISTS permission_migration_probe")
	if err := Migrate(adminDB, ctypes.Runtime_Core, MigrationManifest{
		Runtime: ctypes.Runtime_Core,
		Tables:  []any{&migrationPermissionProbe{}},
	}); err != nil {
		t.Fatalf("migrate through Core role: %v", err)
	}

	var migrationProbeOwner string
	if err := adminDB.Raw(`
		SELECT tableowner
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename = 'permission_migration_probe'
	`).Scan(&migrationProbeOwner).Error; err != nil {
		t.Fatalf("read migration probe owner: %v", err)
	}
	if migrationProbeOwner != ctypes.Runtime_Core.RoleName() {
		t.Fatalf("migration probe owner = %q, want %q", migrationProbeOwner, ctypes.Runtime_Core.RoleName())
	}
	if err := Migrate(adminDB, ctypes.Runtime_Core, MigrationManifest{
		Runtime: ctypes.Runtime_Core,
		Views:   []string{"SELECT FROM"},
	}); err == nil {
		t.Fatal("invalid migration must fail")
	}
	if err := adminDB.Raw("SELECT has_schema_privilege(?, 'public', 'CREATE')", ctypes.Runtime_Core.RoleName()).Scan(&coreCanCreateSchema).Error; err != nil {
		t.Fatalf("read Core schema create access after failed migration: %v", err)
	}
	if coreCanCreateSchema {
		t.Fatal("failed migration must not retain schema CREATE")
	}

	const tableName = "permission_integration_probe"
	const functionName = "permission_integration_function"
	if err := adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName)).Error; err != nil {
		t.Fatalf("clean up probe table: %v", err)
	}
	if err := adminDB.Exec("DROP FUNCTION IF EXISTS " + quoteIdentifier(functionName) + "()").Error; err != nil {
		t.Fatalf("clean up probe function: %v", err)
	}
	if err := adminDB.Exec("CREATE TABLE " + quoteIdentifier(tableName) + " (id integer PRIMARY KEY, value text NOT NULL)").Error; err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	defer adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName))
	if err := adminDB.Exec("CREATE FUNCTION " + quoteIdentifier(functionName) + "() RETURNS integer LANGUAGE SQL AS " + quoteLiteral("SELECT 1")).Error; err != nil {
		t.Fatalf("create probe function: %v", err)
	}
	defer adminDB.Exec("DROP FUNCTION IF EXISTS " + quoteIdentifier(functionName) + "()")

	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_Core,
		Objects: []PermissionObject{
			{
				Type: PermissionObjectType_Database,
				Grants: []PermissionGrant{
					{Runtime: ctypes.Runtime_Core, Privileges: []PermissionPrivilege{PermissionPrivilege_Connect}},
					{Runtime: ctypes.Runtime_DurableJob, Privileges: []PermissionPrivilege{PermissionPrivilege_Connect}},
				},
			},
			{
				Type: PermissionObjectType_Schema,
				Name: "public",
				Grants: []PermissionGrant{
					{Runtime: ctypes.Runtime_Core, Privileges: []PermissionPrivilege{PermissionPrivilege_Usage}},
					{Runtime: ctypes.Runtime_DurableJob, Privileges: []PermissionPrivilege{PermissionPrivilege_Usage}},
				},
			},
			{
				Type: PermissionObjectType_DefaultFunction,
				Name: "public",
			},
			{
				Type: PermissionObjectType_Function,
				Name: functionName,
			},
			{
				Type: PermissionObjectType_Table,
				Name: tableName,
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_Core,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Select,
							PermissionPrivilege_Insert,
							PermissionPrivilege_Update,
							PermissionPrivilege_Delete,
						},
					},
					{
						Runtime:    ctypes.Runtime_DurableJob,
						Privileges: []PermissionPrivilege{PermissionPrivilege_Select},
					},
				},
			},
		},
	}
	if err := ApplyPermissions(adminDB, ctypes.Runtime_Core, manifest); err != nil {
		t.Fatalf("apply permission manifest: %v", err)
	}
	if err := VerifyPermissions(adminDB, ctypes.Runtime_Core, manifest); err != nil {
		t.Fatalf("verify permission manifest: %v", err)
	}
	if err := adminDB.Exec("GRANT INSERT ON TABLE " + quoteIdentifier(tableName) + " TO " + quoteIdentifier(ctypes.Runtime_DurableJob.RoleName())).Error; err != nil {
		t.Fatalf("introduce stale DurableJob grant: %v", err)
	}
	if err := adminDB.Exec("GRANT SELECT ON TABLE " + quoteIdentifier(tableName) + " TO PUBLIC").Error; err != nil {
		t.Fatalf("introduce stale PUBLIC grant: %v", err)
	}
	if err := adminDB.Exec("GRANT EXECUTE ON FUNCTION " + quoteIdentifier(functionName) + "() TO PUBLIC").Error; err != nil {
		t.Fatalf("introduce stale PUBLIC function grant: %v", err)
	}
	if err := adminDB.Exec("ALTER DEFAULT PRIVILEGES FOR ROLE " + quoteIdentifier(ctypes.Runtime_Core.RoleName()) + " IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO PUBLIC").Error; err != nil {
		t.Fatalf("introduce stale PUBLIC function default grant: %v", err)
	}
	if err := ApplyPermissions(adminDB, ctypes.Runtime_Core, manifest); err != nil {
		t.Fatalf("reconcile stale permissions: %v", err)
	}
	if err := VerifyPermissions(adminDB, ctypes.Runtime_Core, manifest); err != nil {
		t.Fatalf("verify reconciled permissions: %v", err)
	}

	var publicTablePrivilegeCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_class tables
		CROSS JOIN LATERAL aclexplode(COALESCE(tables.relacl, acldefault('r', tables.relowner))) privileges
		WHERE tables.oid = ?::regclass AND privileges.grantee = 0
	`, tableName).Scan(&publicTablePrivilegeCount).Error; err != nil {
		t.Fatalf("read PUBLIC table access: %v", err)
	}
	if publicTablePrivilegeCount != 0 {
		t.Fatalf("PUBLIC table privileges = %d, want 0", publicTablePrivilegeCount)
	}

	var publicFunctionPrivilegeCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_proc functions
		CROSS JOIN LATERAL aclexplode(COALESCE(functions.proacl, acldefault('f', functions.proowner))) privileges
		WHERE functions.oid = ?::regprocedure AND privileges.grantee = 0
	`, functionName+"()").Scan(&publicFunctionPrivilegeCount).Error; err != nil {
		t.Fatalf("read PUBLIC function access: %v", err)
	}
	if publicFunctionPrivilegeCount != 0 {
		t.Fatalf("PUBLIC function privileges = %d, want 0", publicFunctionPrivilegeCount)
	}

	var publicDefaultFunctionPrivilegeCount int64
	if err := adminDB.Raw(`
		SELECT COUNT(*)
		FROM pg_default_acl defaultACL
		JOIN pg_roles owners ON owners.oid = defaultACL.defaclrole
		CROSS JOIN LATERAL aclexplode(defaultACL.defaclacl) privileges
		WHERE owners.rolname = ?
			AND defaultACL.defaclnamespace = 0
			AND defaultACL.defaclobjtype = 'f'
			AND privileges.grantee = 0
	`, ctypes.Runtime_Core.RoleName()).Scan(&publicDefaultFunctionPrivilegeCount).Error; err != nil {
		t.Fatalf("read PUBLIC function default access: %v", err)
	}
	if publicDefaultFunctionPrivilegeCount != 0 {
		t.Fatalf("PUBLIC function default privileges = %d, want 0", publicDefaultFunctionPrivilegeCount)
	}

	var durableJobCanInsert bool
	if err := adminDB.Raw("SELECT has_table_privilege(?, ?::regclass, 'INSERT')", ctypes.Runtime_DurableJob.RoleName(), tableName).Scan(&durableJobCanInsert).Error; err != nil {
		t.Fatalf("read DurableJob insert access: %v", err)
	}
	if durableJobCanInsert {
		t.Fatal("DurableJob stale INSERT grant must be revoked")
	}

	coreDB := integrationSQLDB(t, core)
	defer coreDB.Close()
	durableJobDB := integrationSQLDB(t, durableJob)
	defer durableJobDB.Close()

	if _, err := coreDB.Exec("INSERT INTO "+quoteIdentifier(tableName)+" (id, value) VALUES ($1, $2)", 1, "core"); err != nil {
		t.Fatalf("Core insert should be allowed: %v", err)
	}
	var value string
	if err := durableJobDB.QueryRow("SELECT value FROM "+quoteIdentifier(tableName)+" WHERE id = $1", 1).Scan(&value); err != nil {
		t.Fatalf("DurableJob select should be allowed: %v", err)
	}
	if value != "core" {
		t.Fatalf("DurableJob read value = %q, want core", value)
	}
	if _, err := durableJobDB.Exec("INSERT INTO "+quoteIdentifier(tableName)+" (id, value) VALUES ($1, $2)", 2, "durablejob"); err == nil {
		t.Fatal("DurableJob insert should be denied")
	}
	if _, err := durableJobDB.Exec("CREATE TABLE permission_integration_probe_denied (id integer)"); err == nil {
		t.Fatal("DurableJob DDL should be denied")
	}
}

func TestNotificationRuntimePermissionsIntegration(t *testing.T) {
	if os.Getenv("NOTEGIC_RUN_POSTGRES_PERMISSION_INTEGRATION") != "1" {
		t.Skip("set NOTEGIC_RUN_POSTGRES_PERMISSION_INTEGRATION=1 to run PostgreSQL permission integration tests")
	}

	mainAdmin := integrationPostgresConfig(t, "POSTGRES_PERMISSION_ADMIN")
	notificationAdmin := integrationPostgresConfig(t, "POSTGRES_PERMISSION_NOTIFICATION_ADMIN")
	notification := integrationPostgresConfig(t, "POSTGRES_PERMISSION_NOTIFICATION")
	if mainAdmin.Name == notificationAdmin.Name {
		t.Fatal("notification permission integration database must differ from the main database")
	}

	adminDB, err := Connect(notificationAdmin)
	if err != nil {
		t.Fatalf("connect Notification admin database: %v", err)
	}
	defer Disconnect(adminDB)

	if err := EnsureRuntimeRole(adminDB, ctypes.Runtime_Notification, notification.Password); err != nil {
		t.Fatalf("ensure Notification role: %v", err)
	}

	const tableName = "permission_notification_probe"
	if err := adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName)).Error; err != nil {
		t.Fatalf("clean up Notification probe table: %v", err)
	}
	defer adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName))
	if err := adminDB.Exec("CREATE TABLE " + quoteIdentifier(tableName) + " (id integer PRIMARY KEY, value text NOT NULL)").Error; err != nil {
		t.Fatalf("create Notification probe table: %v", err)
	}

	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_Notification,
		Objects: []PermissionObject{
			{
				Type: PermissionObjectType_Database,
				Grants: []PermissionGrant{
					{Runtime: ctypes.Runtime_Notification, Privileges: []PermissionPrivilege{PermissionPrivilege_Connect}},
				},
			},
			{
				Type: PermissionObjectType_Schema,
				Name: "public",
				Grants: []PermissionGrant{
					{Runtime: ctypes.Runtime_Notification, Privileges: []PermissionPrivilege{PermissionPrivilege_Usage}},
				},
			},
			{
				Type: PermissionObjectType_Table,
				Name: tableName,
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_Notification,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Select,
							PermissionPrivilege_Insert,
							PermissionPrivilege_Update,
							PermissionPrivilege_Delete,
						},
					},
				},
			},
		},
	}
	if err := ApplyPermissions(adminDB, ctypes.Runtime_Notification, manifest); err != nil {
		t.Fatalf("apply Notification permission manifest: %v", err)
	}
	if err := VerifyPermissions(adminDB, ctypes.Runtime_Notification, manifest); err != nil {
		t.Fatalf("verify Notification permission manifest: %v", err)
	}

	notificationDB := integrationSQLDB(t, notification)
	defer notificationDB.Close()
	if _, err := notificationDB.Exec("INSERT INTO "+quoteIdentifier(tableName)+" (id, value) VALUES ($1, $2)", 1, "notification"); err != nil {
		t.Fatalf("Notification insert should be allowed: %v", err)
	}
	if _, err := notificationDB.Exec("CREATE TABLE permission_notification_probe_denied (id integer)"); err == nil {
		t.Fatal("Notification DDL should be denied")
	}
}

func integrationPostgresConfig(t *testing.T, prefix string) Config {
	t.Helper()

	config, err := LoadConfig(
		os.Getenv(prefix+"_HOST"),
		os.Getenv(prefix+"_USER"),
		os.Getenv(prefix+"_PASSWORD"),
		os.Getenv(prefix+"_NAME"),
		os.Getenv(prefix+"_PORT"),
	)
	if err != nil {
		t.Fatalf("load %s config: %v", prefix, err)
	}
	return config
}

func integrationSQLDB(t *testing.T, config Config) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", ConnectionString(config))
	if err != nil {
		t.Fatalf("open %s database: %v", config.User, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("ping %s database: %v", config.User, err)
	}
	return db
}
