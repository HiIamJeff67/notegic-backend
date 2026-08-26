package postgres

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
)

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

	const tableName = "permission_integration_probe"
	if err := adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName)).Error; err != nil {
		t.Fatalf("clean up probe table: %v", err)
	}
	if err := adminDB.Exec("CREATE TABLE " + quoteIdentifier(tableName) + " (id integer PRIMARY KEY, value text NOT NULL)").Error; err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	defer adminDB.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(tableName))

	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_Core,
		Objects: []PermissionObject{{
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
		}},
	}
	if err := ApplyPermissions(adminDB, ctypes.Runtime_Core, manifest); err != nil {
		t.Fatalf("apply permission manifest: %v", err)
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
