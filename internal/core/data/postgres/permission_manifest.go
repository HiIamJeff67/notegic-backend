package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

var DatabasePermissionManifest = spostgres.PermissionManifest{
	Runtime: ctypes.Runtime_Core,
	Objects: getCorePermissionObjects(),
}

func getCorePermissionObjects() []spostgres.PermissionObject {
	allTablePrivileges := []spostgres.PermissionPrivilege{
		spostgres.PermissionPrivilege_Select,
		spostgres.PermissionPrivilege_Insert,
		spostgres.PermissionPrivilege_Update,
		spostgres.PermissionPrivilege_Delete,
	}
	durableWritableTables := map[string]bool{
		spostgres.TableName_BlockTable.String():                true,
		spostgres.TableName_BlockPackTable.String():            true,
		spostgres.TableName_BlockPackYjsDocumentTable.String(): true,
		spostgres.TableName_MaterialTable.String():             true,
		spostgres.TableName_SubShelfTable.String():             true,
	}
	durableReadableTables := map[string]bool{
		spostgres.TableName_PlanLimitationTable.String(): true,
	}
	objects := make([]spostgres.PermissionObject, 0, len(DatabaseMigrationManifest.Tables)+len(DatabaseMigrationManifest.Enums)+1)
	for _, table := range DatabaseMigrationManifest.Tables {
		tabler, ok := table.(interface{ TableName() string })
		if !ok {
			continue
		}
		grants := []spostgres.PermissionGrant{{
			Runtime:    ctypes.Runtime_Core,
			Privileges: allTablePrivileges,
		}}
		if durableWritableTables[tabler.TableName()] {
			grants = append(grants, spostgres.PermissionGrant{
				Runtime:    ctypes.Runtime_DurableJob,
				Privileges: allTablePrivileges,
			})
		}
		if durableReadableTables[tabler.TableName()] {
			grants = append(grants, spostgres.PermissionGrant{
				Runtime: ctypes.Runtime_DurableJob,
				Privileges: []spostgres.PermissionPrivilege{
					spostgres.PermissionPrivilege_Select,
				},
			})
		}
		objects = append(objects, spostgres.PermissionObject{
			Type:   spostgres.PermissionObjectType_Table,
			Name:   tabler.TableName(),
			Grants: grants,
		})
	}
	for enumName := range DatabaseMigrationManifest.Enums {
		objects = append(objects, spostgres.PermissionObject{
			Type: spostgres.PermissionObjectType_Enum,
			Name: enumName,
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
				{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
			},
		})
	}
	objects = append(objects, spostgres.PermissionObject{
		Type: spostgres.PermissionObjectType_View,
		Name: spostgres.TableName_UserView.String(),
		Grants: []spostgres.PermissionGrant{
			{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}},
			{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}},
		},
	})
	return objects
}
