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
	if err := db.Exec("ALTER ROLE " + quotedRoleName + " NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS").Error; err != nil {
		return err
	}

	var parentRoleNames []string
	if err := db.Raw(`
		SELECT parent.rolname
		FROM pg_auth_members memberships
		JOIN pg_roles parent ON parent.oid = memberships.roleid
		JOIN pg_roles member ON member.oid = memberships.member
		WHERE member.rolname = ?
		ORDER BY parent.rolname
	`, roleName).Scan(&parentRoleNames).Error; err != nil {
		return err
	}
	for _, parentRoleName := range parentRoleNames {
		if err := db.Exec("REVOKE " + quoteIdentifier(parentRoleName) + " FROM " + quotedRoleName).Error; err != nil {
			return err
		}
	}

	var memberRoleNames []string
	if err := db.Raw(`
		SELECT member.rolname
		FROM pg_auth_members memberships
		JOIN pg_roles parent ON parent.oid = memberships.roleid
		JOIN pg_roles member ON member.oid = memberships.member
		WHERE parent.rolname = ?
		ORDER BY member.rolname
	`, roleName).Scan(&memberRoleNames).Error; err != nil {
		return err
	}
	for _, memberRoleName := range memberRoleNames {
		if err := db.Exec("REVOKE " + quotedRoleName + " FROM " + quoteIdentifier(memberRoleName)).Error; err != nil {
			return err
		}
	}

	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		return err
	}
	if err := db.Exec("REVOKE ALL PRIVILEGES ON DATABASE " + quoteIdentifier(databaseName) + " FROM PUBLIC").Error; err != nil {
		return err
	}
	if err := db.Exec("REVOKE ALL PRIVILEGES ON SCHEMA public FROM PUBLIC").Error; err != nil {
		return err
	}
	if err := db.Exec("GRANT CONNECT ON DATABASE " + quoteIdentifier(databaseName) + " TO " + quotedRoleName).Error; err != nil {
		return err
	}
	return db.Exec("GRANT USAGE ON SCHEMA public TO " + quotedRoleName).Error
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
		objectSQL := ""
		switch object.Type {
		case PermissionObjectType_Database:
			var databaseName string
			if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
				return err
			}
			objectSQL = "DATABASE " + quoteIdentifier(databaseName)
		case PermissionObjectType_Schema:
			objectSQL = "SCHEMA " + quoteIdentifier(object.Name)
		case PermissionObjectType_DefaultTable, PermissionObjectType_DefaultSequence, PermissionObjectType_DefaultFunction:
			defaultObjectSQL := ""
			switch object.Type {
			case PermissionObjectType_DefaultTable:
				defaultObjectSQL = "TABLES"
			case PermissionObjectType_DefaultSequence:
				defaultObjectSQL = "SEQUENCES"
			case PermissionObjectType_DefaultFunction:
				defaultObjectSQL = "FUNCTIONS"
			}
			globalPrefix := "ALTER DEFAULT PRIVILEGES FOR ROLE " + quoteIdentifier(manifest.Runtime.RoleName())
			schemaPrefix := globalPrefix + " IN SCHEMA " + quoteIdentifier(object.Name)
			if err := db.Exec(globalPrefix + " REVOKE ALL PRIVILEGES ON " + defaultObjectSQL + " FROM PUBLIC").Error; err != nil {
				return err
			}
			for candidate, exists := range roles {
				if !exists {
					continue
				}
				if err := db.Exec(globalPrefix + " REVOKE ALL PRIVILEGES ON " + defaultObjectSQL + " FROM " + quoteIdentifier(candidate.RoleName())).Error; err != nil {
					return err
				}
				if err := db.Exec(schemaPrefix + " REVOKE ALL PRIVILEGES ON " + defaultObjectSQL + " FROM " + quoteIdentifier(candidate.RoleName())).Error; err != nil {
					return err
				}
			}
			if err := db.Exec(schemaPrefix + " REVOKE ALL PRIVILEGES ON " + defaultObjectSQL + " FROM PUBLIC").Error; err != nil {
				return err
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
				if err := db.Exec(schemaPrefix + " GRANT " + strings.Join(privileges, ", ") + " ON " + defaultObjectSQL + " TO " + quoteIdentifier(grant.Runtime.RoleName())).Error; err != nil {
					return err
				}
			}
			continue
		case PermissionObjectType_Table, PermissionObjectType_View:
			objectSQL = "TABLE " + quoteIdentifier(object.Name)
		case PermissionObjectType_Enum:
			objectSQL = "TYPE " + quoteIdentifier(object.Name)
		case PermissionObjectType_Sequence:
			objectSQL = "SEQUENCE " + quoteIdentifier(object.Name)
		case PermissionObjectType_Function:
			objectSQL = "FUNCTION " + quoteIdentifier(object.Name) + "()"
		default:
			return fmt.Errorf("unsupported permission object type %q", object.Type)
		}
		if err := db.Exec("REVOKE ALL PRIVILEGES ON " + objectSQL + " FROM PUBLIC").Error; err != nil {
			return err
		}
		for candidate, exists := range roles {
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

// VerifyPermissions checks that each non-owner runtime has exactly the
// privileges declared by a manifest. Owner privileges are implicit in
// PostgreSQL object ownership and are therefore intentionally not compared.
func VerifyPermissions(db *gorm.DB, runtime ctypes.Runtime, manifest PermissionManifest) error {
	if db == nil {
		return fmt.Errorf("admin database connection is required")
	}
	if !manifest.IsFor(runtime) {
		return fmt.Errorf("runtime %q cannot verify permissions for runtime %q", runtime, manifest.Runtime)
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
		if object.Type == PermissionObjectType_DefaultTable || object.Type == PermissionObjectType_DefaultSequence || object.Type == PermissionObjectType_DefaultFunction {
			defaultObjectType := ""
			switch object.Type {
			case PermissionObjectType_DefaultTable:
				defaultObjectType = "r"
			case PermissionObjectType_DefaultSequence:
				defaultObjectType = "S"
			case PermissionObjectType_DefaultFunction:
				defaultObjectType = "f"
			}

			for _, candidate := range ctypes.AllRuntimes {
				if !roles[candidate] || candidate == manifest.Runtime {
					continue
				}
				for _, privilege := range object.Type.Privileges() {
					expected := false
					for _, grant := range object.Grants {
						if grant.Runtime != candidate {
							continue
						}
						for _, grantedPrivilege := range grant.Privileges {
							if grantedPrivilege == privilege {
								expected = true
							}
						}
					}

					var actual bool
					if err := db.Raw(`
						WITH owner_role AS (
							SELECT oid
							FROM pg_roles
							WHERE rolname = ?
						), default_acls AS (
							SELECT defaultACL.defaclacl
							FROM pg_default_acl defaultACL
							JOIN owner_role ON owner_role.oid = defaultACL.defaclrole
							WHERE defaultACL.defaclobjtype = ?
								AND defaultACL.defaclnamespace = 0
							UNION ALL
							SELECT defaultACL.defaclacl
							FROM pg_default_acl defaultACL
							JOIN owner_role ON owner_role.oid = defaultACL.defaclrole
							JOIN pg_namespace schemas ON schemas.oid = defaultACL.defaclnamespace
							WHERE defaultACL.defaclobjtype = ?
								AND schemas.nspname = ?
							UNION ALL
							SELECT acldefault(?, owner_role.oid)
							FROM owner_role
							WHERE NOT EXISTS (
								SELECT 1
								FROM pg_default_acl defaultACL
								WHERE defaultACL.defaclrole = owner_role.oid
									AND defaultACL.defaclnamespace = 0
									AND defaultACL.defaclobjtype = ?
							)
						)
						SELECT EXISTS (
							SELECT 1
							FROM default_acls
							CROSS JOIN LATERAL aclexplode(default_acls.defaclacl) privileges
							JOIN pg_roles grantees ON grantees.oid = privileges.grantee
							WHERE grantees.rolname = ? AND privileges.privilege_type = ?
						)
					`, manifest.Runtime.RoleName(), defaultObjectType, defaultObjectType, object.Name, defaultObjectType, defaultObjectType, candidate.RoleName(), string(privilege)).Scan(&actual).Error; err != nil {
						return err
					}
					if actual != expected {
						return fmt.Errorf("permission drift: runtime %q has %t %s on %s %q, want %t", candidate, actual, privilege, object.Type, object.Name, expected)
					}
				}
			}

			for _, privilege := range object.Type.Privileges() {
				var publicActual bool
				if err := db.Raw(`
					WITH owner_role AS (
						SELECT oid
						FROM pg_roles
						WHERE rolname = ?
					), default_acls AS (
						SELECT defaultACL.defaclacl
						FROM pg_default_acl defaultACL
						JOIN owner_role ON owner_role.oid = defaultACL.defaclrole
						WHERE defaultACL.defaclobjtype = ?
							AND defaultACL.defaclnamespace = 0
						UNION ALL
						SELECT defaultACL.defaclacl
						FROM pg_default_acl defaultACL
						JOIN owner_role ON owner_role.oid = defaultACL.defaclrole
						JOIN pg_namespace schemas ON schemas.oid = defaultACL.defaclnamespace
						WHERE defaultACL.defaclobjtype = ?
							AND schemas.nspname = ?
						UNION ALL
						SELECT acldefault(?, owner_role.oid)
						FROM owner_role
						WHERE NOT EXISTS (
							SELECT 1
							FROM pg_default_acl defaultACL
							WHERE defaultACL.defaclrole = owner_role.oid
								AND defaultACL.defaclnamespace = 0
								AND defaultACL.defaclobjtype = ?
						)
					)
					SELECT EXISTS (
						SELECT 1
						FROM default_acls
						CROSS JOIN LATERAL aclexplode(default_acls.defaclacl) privileges
						WHERE privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, manifest.Runtime.RoleName(), defaultObjectType, defaultObjectType, object.Name, defaultObjectType, defaultObjectType, string(privilege)).Scan(&publicActual).Error; err != nil {
					return err
				}
				if publicActual {
					return fmt.Errorf("permission drift: PUBLIC has %s on %s %q", privilege, object.Type, object.Name)
				}
			}

			continue
		}

		for _, privilege := range object.Type.Privileges() {
			var publicActual bool
			var err error
			switch object.Type {
			case PermissionObjectType_Database:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_database databases
						CROSS JOIN LATERAL aclexplode(COALESCE(databases.datacl, acldefault('d', databases.datdba))) privileges
						WHERE databases.datname = current_database() AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, string(privilege)).Scan(&publicActual).Error
			case PermissionObjectType_Schema:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_namespace schemas
						CROSS JOIN LATERAL aclexplode(COALESCE(schemas.nspacl, acldefault('n', schemas.nspowner))) privileges
						WHERE schemas.nspname = ? AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, object.Name, string(privilege)).Scan(&publicActual).Error
			case PermissionObjectType_Table, PermissionObjectType_View:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_class tables
						JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
						CROSS JOIN LATERAL aclexplode(COALESCE(tables.relacl, acldefault('r', tables.relowner))) privileges
						WHERE schemas.nspname = 'public' AND tables.relname = ? AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, object.Name, string(privilege)).Scan(&publicActual).Error
			case PermissionObjectType_Enum:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_type types
						JOIN pg_namespace schemas ON schemas.oid = types.typnamespace
						CROSS JOIN LATERAL aclexplode(COALESCE(types.typacl, acldefault('T', types.typowner))) privileges
						WHERE schemas.nspname = 'public' AND types.typname = ? AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, object.Name, string(privilege)).Scan(&publicActual).Error
			case PermissionObjectType_Sequence:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_class sequences
						JOIN pg_namespace schemas ON schemas.oid = sequences.relnamespace
						CROSS JOIN LATERAL aclexplode(COALESCE(sequences.relacl, acldefault('S', sequences.relowner))) privileges
						WHERE schemas.nspname = 'public' AND sequences.relname = ? AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, object.Name, string(privilege)).Scan(&publicActual).Error
			case PermissionObjectType_Function:
				err = db.Raw(`
					SELECT EXISTS (
						SELECT 1
						FROM pg_proc functions
						CROSS JOIN LATERAL aclexplode(COALESCE(functions.proacl, acldefault('f', functions.proowner))) privileges
						WHERE functions.oid = ?::regprocedure AND privileges.grantee = 0 AND privileges.privilege_type = ?
					)
				`, object.Name+"()", string(privilege)).Scan(&publicActual).Error
			}
			if err != nil {
				return err
			}
			if publicActual {
				return fmt.Errorf("permission drift: PUBLIC has %s on %s %q", privilege, object.Type, object.Name)
			}
		}

		for _, candidate := range ctypes.AllRuntimes {
			if !roles[candidate] {
				continue
			}
			if candidate == manifest.Runtime && object.Type != PermissionObjectType_Database && object.Type != PermissionObjectType_Schema {
				continue
			}
			for _, privilege := range object.Type.Privileges() {
				expected := false
				for _, grant := range object.Grants {
					if grant.Runtime != candidate {
						continue
					}
					for _, grantedPrivilege := range grant.Privileges {
						if grantedPrivilege == privilege {
							expected = true
						}
					}
				}
				var actual bool
				var err error
				switch object.Type {
				case PermissionObjectType_Database:
					err = db.Raw("SELECT has_database_privilege(?, current_database(), ?)", candidate.RoleName(), string(privilege)).Scan(&actual).Error
				case PermissionObjectType_Schema:
					err = db.Raw("SELECT has_schema_privilege(?, ?, ?)", candidate.RoleName(), object.Name, string(privilege)).Scan(&actual).Error
				case PermissionObjectType_Table, PermissionObjectType_View:
					err = db.Raw(`
						SELECT EXISTS (
							SELECT 1
							FROM pg_class tables
							JOIN pg_namespace schemas ON schemas.oid = tables.relnamespace
							WHERE schemas.nspname = 'public'
								AND tables.relname = ?
								AND has_table_privilege(?, tables.oid, ?)
						)
					`, object.Name, candidate.RoleName(), string(privilege)).Scan(&actual).Error
				case PermissionObjectType_Enum:
					err = db.Raw(`
						SELECT EXISTS (
							SELECT 1
							FROM pg_type types
							JOIN pg_namespace schemas ON schemas.oid = types.typnamespace
							WHERE schemas.nspname = 'public'
								AND types.typname = ?
								AND has_type_privilege(?, types.oid, ?)
						)
					`, object.Name, candidate.RoleName(), string(privilege)).Scan(&actual).Error
				case PermissionObjectType_Sequence:
					err = db.Raw(`
						SELECT EXISTS (
							SELECT 1
							FROM pg_class sequences
							JOIN pg_namespace schemas ON schemas.oid = sequences.relnamespace
							WHERE schemas.nspname = 'public'
								AND sequences.relname = ?
								AND has_sequence_privilege(?, sequences.oid, ?)
						)
					`, object.Name, candidate.RoleName(), string(privilege)).Scan(&actual).Error
				case PermissionObjectType_Function:
					err = db.Raw("SELECT has_function_privilege(?, ?::regprocedure, ?)", candidate.RoleName(), object.Name+"()", string(privilege)).Scan(&actual).Error
				}
				if err != nil {
					return err
				}
				if actual != expected {
					return fmt.Errorf("permission drift: runtime %q has %t %s on %s %q, want %t", candidate, actual, privilege, object.Type, object.Name, expected)
				}
			}
		}
	}

	return nil
}
