package postgres

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
)

func EnsureRuntimeRole(db *gorm.DB, runtime ctypes.Runtime, password string) error {
	if db == nil {
		return fmt.Errorf("admin database connection is required")
	}
	if !runtime.IsValid() {
		return fmt.Errorf("invalid runtime %q", runtime)
	}
	if password == "" {
		return fmt.Errorf("password for runtime %q is required", runtime)
	}

	roleName := runtime.RoleName()
	quotedRoleName := quoteIdentifier(roleName)
	var roleExists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = ?)", roleName).Scan(&roleExists).Error; err != nil {
		return err
	}

	roleStatement := "CREATE ROLE " + quotedRoleName + " LOGIN PASSWORD " + quoteLiteral(password)
	if roleExists {
		roleStatement = "ALTER ROLE " + quotedRoleName + " LOGIN PASSWORD " + quoteLiteral(password)
	}
	if err := db.Exec(roleStatement).Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER ROLE " + quotedRoleName + " NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS").Error; err != nil {
		return err
	}

	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		return err
	}
	if err := db.Exec("GRANT CONNECT ON DATABASE " + quoteIdentifier(databaseName) + " TO " + quotedRoleName).Error; err != nil {
		return err
	}
	return db.Exec("GRANT USAGE ON SCHEMA public TO " + quotedRoleName).Error
}

func GrantMigrationAccess(db *gorm.DB, runtime ctypes.Runtime) error {
	if db == nil {
		return fmt.Errorf("admin database connection is required")
	}
	if !runtime.IsValid() {
		return fmt.Errorf("invalid runtime %q", runtime)
	}

	return db.Exec("GRANT CREATE ON SCHEMA public TO " + quoteIdentifier(runtime.RoleName())).Error
}

func RevokeMigrationAccess(db *gorm.DB, runtime ctypes.Runtime) error {
	if db == nil {
		return fmt.Errorf("admin database connection is required")
	}
	if !runtime.IsValid() {
		return fmt.Errorf("invalid runtime %q", runtime)
	}

	return db.Exec("REVOKE CREATE ON SCHEMA public FROM " + quoteIdentifier(runtime.RoleName())).Error
}

func ApplyPermissions(db *gorm.DB, runtime ctypes.Runtime, manifest PermissionManifest) error {
	if db == nil {
		return fmt.Errorf("admin database connection is required")
	}
	if !manifest.IsFor(runtime) {
		return fmt.Errorf("runtime %q cannot apply permissions for runtime %q", runtime, manifest.Runtime)
	}
	if err := manifest.Validate(); err != nil {
		return err
	}

	roles := make(map[ctypes.Runtime]bool, len(ctypes.AllRuntimes))
	for _, candidate := range ctypes.AllRuntimes {
		var exists bool
		if err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = ?)", candidate.RoleName()).Scan(&exists).Error; err != nil {
			return err
		}
		roles[candidate] = exists
	}

	for _, object := range manifest.Objects {
		if strings.TrimSpace(object.Name) == "" {
			return fmt.Errorf("permission object name is required")
		}
		objectName := quoteIdentifier(object.Name)
		objectSQL := ""
		switch object.Type {
		case PermissionObjectType_Table, PermissionObjectType_View:
			objectSQL = "TABLE " + objectName
		case PermissionObjectType_Enum:
			objectSQL = "TYPE " + objectName
		case PermissionObjectType_Sequence:
			objectSQL = "SEQUENCE " + objectName
		case PermissionObjectType_Function:
			objectSQL = "FUNCTION " + objectName
		default:
			return fmt.Errorf("unsupported permission object type %q", object.Type)
		}
		grantRuntimes := make(map[ctypes.Runtime]struct{}, len(object.Grants))
		for _, grant := range object.Grants {
			grantRuntimes[grant.Runtime] = struct{}{}
		}
		for candidate := range grantRuntimes {
			exists := roles[candidate]
			if !exists {
				continue
			}
			if err := db.Exec("REVOKE ALL PRIVILEGES ON " + objectSQL + " FROM " + quoteIdentifier(candidate.RoleName())).Error; err != nil {
				return err
			}
		}
		for _, grant := range object.Grants {
			if !roles[grant.Runtime] {
				if grant.Runtime == runtime {
					return fmt.Errorf("runtime role %q does not exist", grant.Runtime.RoleName())
				}
				continue
			}
			privileges := make([]string, len(grant.Privileges))
			for index, privilege := range grant.Privileges {
				privileges[index] = string(privilege)
			}
			statement := "GRANT " + strings.Join(privileges, ", ") + " ON " + objectSQL + " TO " + quoteIdentifier(grant.Runtime.RoleName())
			if err := db.Exec(statement).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
