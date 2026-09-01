package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	saccountingtrigger "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/triggers/accounting_triggers"
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
	routineTaskPrivileges := append(
		append([]spostgres.PermissionPrivilege{}, privileges...),
		spostgres.PermissionPrivilege_Trigger,
	)
	objects := []spostgres.PermissionObject{
		{
			Type: spostgres.PermissionObjectType_Database,
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Connect}},
				{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Connect}},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Schema,
			Name: "public",
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
				{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
			},
		},
		{
			Type: spostgres.PermissionObjectType_DefaultFunction,
			Name: "public",
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_RoutineDependencyTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges},
				{Runtime: ctypes.Runtime_Core, Privileges: privileges},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_RoutineRecordTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: privileges},
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{
					spostgres.PermissionPrivilege_Select,
				}},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Table,
			Name: spostgres.TableName_RoutineTaskTable.String(),
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_DurableJob, Privileges: routineTaskPrivileges},
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
	// DurableJob owns these trigger functions, while Core needs EXECUTE when
	// updating the RoutineTask table whose triggers invoke them.
	for _, functionName := range []string{
		saccountingtrigger.AccountingInsertedRoutineTaskTriggerFunctionName,
		saccountingtrigger.AccountingUpdatedRoutineTaskTriggerFunctionName,
	} {
		objects = append(objects, spostgres.PermissionObject{
			Type: spostgres.PermissionObjectType_Function,
			Name: functionName,
			Grants: []spostgres.PermissionGrant{
				{
					Runtime:    ctypes.Runtime_Core,
					Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Execute},
				},
				{
					Runtime:    ctypes.Runtime_DurableJob,
					Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Execute},
				},
			},
		})
	}
	return objects
}
