package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

var DatabasePermissionManifest = spostgres.PermissionManifest{
	Runtime: ctypes.Runtime_DurableJob,
	Objects: getDurablePermissionObjects(),
}

func getDurablePermissionObjects() []spostgres.PermissionObject {
	// DurableJob and Core both need complete DML access to routine-task and
	// quota tables because they coordinate task state and quota accounting.
	privileges := []spostgres.PermissionPrivilege{
		spostgres.PermissionPrivilege_Select,
		spostgres.PermissionPrivilege_Insert,
		spostgres.PermissionPrivilege_Update,
		spostgres.PermissionPrivilege_Delete,
	}
	objects := []spostgres.PermissionObject{
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_RoutineTaskTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges},
				{Runtime: ctypes.Runtime_Core, Privileges: privileges},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_RoutineTaskRecordTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges},
				{Runtime: ctypes.Runtime_Core, Privileges: privileges},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_UserQuotaTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges},
				{Runtime: ctypes.Runtime_Core, Privileges: privileges},
			},
		},
	}
	for _, tableName := range []string{
		spostgres.TableName_BlockTable.String(),
		spostgres.TableName_BlockPackTable.String(),
		spostgres.TableName_BlockPackYjsDocumentTable.String(),
		spostgres.TableName_MaterialTable.String(),
		spostgres.TableName_SubShelfTable.String(),
	} {
		objects = append(objects, spostgres.PermissionObject{
			Type:   spostgres.PermissionObjectType_Table,
			Name:   tableName,
			Grants: []spostgres.PermissionGrant{{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges}},
		})
	}
	objects = append(objects,
		spostgres.PermissionObject{
			Type:   spostgres.PermissionObjectType_Table,
			Name:   spostgres.TableName_PlanLimitationTable.String(),
			Grants: []spostgres.PermissionGrant{{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}}},
		},
		spostgres.PermissionObject{
			Type:   spostgres.PermissionObjectType_View,
			Name:   spostgres.TableName_UserView.String(),
			Grants: []spostgres.PermissionGrant{{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}}},
		},
	)
	for _, enumName := range []string{
		new(cenums.AccessControlPermission).Name(),
		new(cenums.BadgeType).Name(),
		new(cenums.BillingIntervalUnit).Name(),
		new(cenums.BillingPlanName).Name(),
		new(cenums.BillingPlanStatus).Name(),
		new(cenums.BlockType).Name(),
		new(cenums.CountryCode).Name(),
		new(cenums.Country).Name(),
		new(cenums.ItemType).Name(),
		new(cenums.Language).Name(),
		new(cenums.UserSettingDensity).Name(),
		new(cenums.UserSettingStartSurface).Name(),
		new(cenums.MaterialContentType).Name(),
		new(cenums.RoutinePeriod).Name(),
		new(cenums.RoutineStatus).Name(),
		new(cenums.RoutineTaskPurpose).Name(),
		new(cenums.RoutineTaskStatus).Name(),
		new(cenums.RoutineTaskRecordErrorCode).Name(),
		new(cenums.RoutineTaskRecordStatus).Name(),
		new(cenums.SupportedIcon).Name(),
		new(cenums.SupportedCurrencyCode).Name(),
		new(cenums.UserGender).Name(),
		new(cenums.UserPlan).Name(),
		new(cenums.UserRole).Name(),
		new(cenums.UserStatus).Name(),
		new(cenums.UsersToBillingPlansStatus).Name(),
	} {
		objects = append(objects, spostgres.PermissionObject{
			Type:   spostgres.PermissionObjectType_Enum,
			Name:   enumName,
			Grants: []spostgres.PermissionGrant{{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}}},
		})
	}
	return objects
}
