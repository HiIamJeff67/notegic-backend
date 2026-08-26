package postgres

import (
	"fmt"
	"slices"
	"strings"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
)

// PermissionObjectType identifies the PostgreSQL object class being granted.
// Triggers and constraints are intentionally absent because their authority
// follows the table that owns them.
type PermissionObjectType string

const (
	PermissionObjectType_Table    PermissionObjectType = "table"
	PermissionObjectType_View     PermissionObjectType = "view"
	PermissionObjectType_Enum     PermissionObjectType = "enum"
	PermissionObjectType_Sequence PermissionObjectType = "sequence"
	PermissionObjectType_Function PermissionObjectType = "function"
)

func (objectType PermissionObjectType) IsValid() bool {
	return slices.Contains([]PermissionObjectType{
		PermissionObjectType_Table,
		PermissionObjectType_View,
		PermissionObjectType_Enum,
		PermissionObjectType_Sequence,
		PermissionObjectType_Function,
	}, objectType)
}

func (objectType PermissionObjectType) Allows(privilege PermissionPrivilege) bool {
	switch objectType {
	case PermissionObjectType_Table:
		return slices.Contains([]PermissionPrivilege{
			PermissionPrivilege_Select,
			PermissionPrivilege_Insert,
			PermissionPrivilege_Update,
			PermissionPrivilege_Delete,
		}, privilege)
	case PermissionObjectType_View:
		return privilege == PermissionPrivilege_Select
	case PermissionObjectType_Enum:
		return privilege == PermissionPrivilege_Usage
	case PermissionObjectType_Sequence:
		return slices.Contains([]PermissionPrivilege{
			PermissionPrivilege_Usage,
			PermissionPrivilege_Select,
		}, privilege)
	case PermissionObjectType_Function:
		return privilege == PermissionPrivilege_Execute
	}

	return false
}

// PermissionPrivilege identifies a PostgreSQL object privilege.
type PermissionPrivilege string

const (
	PermissionPrivilege_Select  PermissionPrivilege = "SELECT"
	PermissionPrivilege_Insert  PermissionPrivilege = "INSERT"
	PermissionPrivilege_Update  PermissionPrivilege = "UPDATE"
	PermissionPrivilege_Delete  PermissionPrivilege = "DELETE"
	PermissionPrivilege_Usage   PermissionPrivilege = "USAGE"
	PermissionPrivilege_Execute PermissionPrivilege = "EXECUTE"
)

// PermissionGrant describes one runtime role and its privileges on an object.
type PermissionGrant struct {
	Runtime    ctypes.Runtime
	Privileges []PermissionPrivilege
}

// PermissionObject describes one named PostgreSQL object and the runtime
// grants that should eventually be reconciled for it.
type PermissionObject struct {
	Type   PermissionObjectType
	Name   string
	Grants []PermissionGrant
}

// PermissionManifest describes the PostgreSQL privileges to reconcile during a
// runtime bootstrap. Objects may be owned by another runtime's migration
// manifest. It is intentionally only a declaration; it does not execute SQL.
type PermissionManifest struct {
	Runtime ctypes.Runtime
	Objects []PermissionObject
}

func (manifest PermissionManifest) IsFor(runtime ctypes.Runtime) bool {
	return manifest.Runtime == runtime
}

func (manifest PermissionManifest) Validate() error {
	if !manifest.Runtime.IsValid() {
		return fmt.Errorf("invalid permission manifest runtime %q", manifest.Runtime)
	}

	seenObjects := make(map[string]struct{}, len(manifest.Objects))
	for _, object := range manifest.Objects {
		if !object.Type.IsValid() {
			return fmt.Errorf("invalid permission object type %q", object.Type)
		}
		if strings.TrimSpace(object.Name) == "" {
			return fmt.Errorf("permission object name is required")
		}

		objectKey := string(object.Type) + ":" + object.Name
		if _, exists := seenObjects[objectKey]; exists {
			return fmt.Errorf("duplicate permission object %q", objectKey)
		}
		seenObjects[objectKey] = struct{}{}

		seenGrantees := make(map[ctypes.Runtime]struct{}, len(object.Grants))
		for _, grant := range object.Grants {
			if !grant.Runtime.IsValid() {
				return fmt.Errorf("invalid permission grant runtime %q", grant.Runtime)
			}
			if _, exists := seenGrantees[grant.Runtime]; exists {
				return fmt.Errorf("duplicate permission grant runtime %q for %q", grant.Runtime, objectKey)
			}
			seenGrantees[grant.Runtime] = struct{}{}
			if len(grant.Privileges) == 0 {
				return fmt.Errorf("permission grant for %q has no privileges", objectKey)
			}

			seenPrivileges := make(map[PermissionPrivilege]struct{}, len(grant.Privileges))
			for _, privilege := range grant.Privileges {
				if !object.Type.Allows(privilege) {
					return fmt.Errorf("permission %q is not valid for object type %q", privilege, object.Type)
				}
				if _, exists := seenPrivileges[privilege]; exists {
					return fmt.Errorf("duplicate permission %q for %q", privilege, objectKey)
				}
				seenPrivileges[privilege] = struct{}{}
			}
		}
	}

	return nil
}
