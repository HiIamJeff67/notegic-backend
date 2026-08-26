package postgres

import (
	"testing"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
)

func TestPermissionManifestValidate(t *testing.T) {
	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_Core,
		Objects: []PermissionObject{
			{
				Type: PermissionObjectType_Table,
				Name: "UserTable",
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
				},
			},
			{
				Type: PermissionObjectType_View,
				Name: "UserView",
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_DurableJob,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Select,
						},
					},
				},
			},
		},
	}

	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !manifest.IsFor(ctypes.Runtime_Core) {
		t.Fatal("IsFor() should match the manifest runtime")
	}
}

func TestPermissionManifestValidateRejectsInvalidPrivilege(t *testing.T) {
	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_Notification,
		Objects: []PermissionObject{
			{
				Type: PermissionObjectType_View,
				Name: "UserView",
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_Notification,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Insert,
						},
					},
				},
			},
		},
	}

	if err := manifest.Validate(); err == nil {
		t.Fatal("Validate() expected an invalid privilege error")
	}
}

func TestPermissionManifestValidateRejectsDuplicateObjects(t *testing.T) {
	manifest := PermissionManifest{
		Runtime: ctypes.Runtime_DurableJob,
		Objects: []PermissionObject{
			{
				Type: PermissionObjectType_Enum,
				Name: "RoutineTaskStatus",
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_DurableJob,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Usage,
						},
					},
				},
			},
			{
				Type: PermissionObjectType_Enum,
				Name: "RoutineTaskStatus",
				Grants: []PermissionGrant{
					{
						Runtime: ctypes.Runtime_DurableJob,
						Privileges: []PermissionPrivilege{
							PermissionPrivilege_Usage,
						},
					},
				},
			},
		},
	}

	if err := manifest.Validate(); err == nil {
		t.Fatal("Validate() expected a duplicate object error")
	}
}
