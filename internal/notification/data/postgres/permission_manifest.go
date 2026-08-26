package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

var DatabasePermissionManifest = spostgres.PermissionManifest{
	Runtime: ctypes.Runtime_Notification,
	Objects: []spostgres.PermissionObject{
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_NotificationTable.String(),
			Grants: []spostgres.PermissionGrant{
				{
					Runtime: ctypes.Runtime_Notification,
					// Notification owns both tables and needs complete DML access
					// to persist notifications and maintain the user projection.
					Privileges: []spostgres.PermissionPrivilege{
						spostgres.PermissionPrivilege_Select,
						spostgres.PermissionPrivilege_Insert,
						spostgres.PermissionPrivilege_Update,
						spostgres.PermissionPrivilege_Delete,
					},
				},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_UserProjection.String(),
			Grants: []spostgres.PermissionGrant{
				{
					Runtime: ctypes.Runtime_Notification,
					Privileges: []spostgres.PermissionPrivilege{
						spostgres.PermissionPrivilege_Select,
						spostgres.PermissionPrivilege_Insert,
						spostgres.PermissionPrivilege_Update,
						spostgres.PermissionPrivilege_Delete,
					},
				},
			},
		},
	},
}
