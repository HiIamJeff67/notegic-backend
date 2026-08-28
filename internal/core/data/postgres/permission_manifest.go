package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	saccountingtrigger "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers/accounting_triggers"
	sblockpackyjstrigger "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers/block_pack_yjs_triggers"
	sitemstrigger "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers/item_projection_triggers"
	sshelfitemcascadingtrigger "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers/shelf_item_cascading_triggers"
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
	coreTablePrivileges := append(append([]spostgres.PermissionPrivilege{}, allTablePrivileges...), spostgres.PermissionPrivilege_Trigger)
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
	yjsReadableTables := map[string]bool{
		spostgres.TableName_BlockTable.String():                true,
		spostgres.TableName_BlockPackTable.String():            true,
		spostgres.TableName_BlockPackYjsDocumentTable.String(): true,
		spostgres.TableName_BlockPackYjsUpdateTable.String():   true,
	}
	yjsWritableTables := map[string]bool{
		spostgres.TableName_BlockTable.String():                true,
		spostgres.TableName_BlockPackYjsDocumentTable.String(): true,
		spostgres.TableName_BlockPackYjsUpdateTable.String():   true,
	}
	// Yjs writes projected blocks directly. The Block accounting triggers are
	// SECURITY DEFINER functions owned by Core, so their quota/accounting table
	// access is not exposed to the Yjs worker role.
	yjsTriggerFunctions := map[string]bool{
		saccountingtrigger.AccountingInsertedBlockTriggerFunctionName: true,
		saccountingtrigger.AccountingDeletedBlockTriggerFunctionName:  true,
	}
	objects := []spostgres.PermissionObject{
		{
			Type: spostgres.PermissionObjectType_Database,
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Connect}},
				{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Connect}},
				{Runtime: ctypes.Runtime_YjsWorker, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Connect}},
			},
		},
		{
			Type: spostgres.PermissionObjectType_Schema,
			Name: "public",
			Grants: []spostgres.PermissionGrant{
				{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
				{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
				{Runtime: ctypes.Runtime_YjsWorker, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage}},
			},
		},
		{
			Type: spostgres.PermissionObjectType_DefaultFunction,
			Name: "public",
		},
	}
	for _, table := range DatabaseMigrationManifest.Tables {
		tabler, ok := table.(interface{ TableName() string })
		if !ok {
			continue
		}
		grants := []spostgres.PermissionGrant{{
			Runtime:    ctypes.Runtime_Core,
			Privileges: coreTablePrivileges,
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
		if yjsReadableTables[tabler.TableName()] {
			privileges := []spostgres.PermissionPrivilege{
				spostgres.PermissionPrivilege_Select,
			}
			if yjsWritableTables[tabler.TableName()] {
				privileges = append(privileges,
					spostgres.PermissionPrivilege_Update,
				)
			}
			if tabler.TableName() == spostgres.TableName_BlockTable.String() {
				privileges = append(privileges,
					spostgres.PermissionPrivilege_Insert,
					spostgres.PermissionPrivilege_Delete,
				)
			}
			grants = append(grants, spostgres.PermissionGrant{
				Runtime:    ctypes.Runtime_YjsWorker,
				Privileges: privileges,
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
		if enumName == new(cenums.BlockType).Name() {
			objects[len(objects)-1].Grants = append(objects[len(objects)-1].Grants,
				spostgres.PermissionGrant{
					Runtime:    ctypes.Runtime_YjsWorker,
					Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Usage},
				},
			)
		}
	}
	objects = append(objects, spostgres.PermissionObject{
		Type: spostgres.PermissionObjectType_View,
		Name: spostgres.TableName_UserView.String(),
		Grants: []spostgres.PermissionGrant{
			{Runtime: ctypes.Runtime_Core, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}},
			{Runtime: ctypes.Runtime_DurableJob, Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Select}},
		},
	})
	for _, functionName := range []string{
		sshelfitemcascadingtrigger.CascadingSoftDeleteRootShelfTriggerFunctionName,
		sshelfitemcascadingtrigger.CascadingSoftDeleteSubShelfTriggerFunctionName,
		sshelfitemcascadingtrigger.CascadingRestoreRootShelfTriggerFunctionName,
		sshelfitemcascadingtrigger.CascadingRestoreSubShelfTriggerFunctionName,
		sshelfitemcascadingtrigger.CascadingMoveSubShelfTriggerFunctionName,
		sblockpackyjstrigger.SyncBlockPackYjsDocumentDeletedAtTriggerFunctionName,
		sitemstrigger.ProjectSubShelvesToItemsTriggerFunctionName,
		sitemstrigger.ProjectMaterialsToItemsTriggerFunctionName,
		sitemstrigger.DeleteMaterialItemsAfterDeleteTriggerFunctionName,
		sitemstrigger.ProjectBlockPacksToItemsTriggerFunctionName,
		sitemstrigger.DeleteBlockPackItemsAfterDeleteTriggerFunctionName,
		saccountingtrigger.AccountingMutatedBlockPackTriggerFunctionName,
		saccountingtrigger.AccountingInsertedBlockTriggerFunctionName,
		saccountingtrigger.AccountingDeletedBlockTriggerFunctionName,
		saccountingtrigger.AccountingMutatedRootShelfTriggerFunctionName,
		saccountingtrigger.AccountingMutatedSubShelfTriggerFunctionName,
		saccountingtrigger.AccountingMutatedMaterialTriggerFunctionName,
		saccountingtrigger.AccountingInsertedRoutineTagTriggerFunctionName,
		saccountingtrigger.AccountingDeletedRoutineTagTriggerFunctionName,
		saccountingtrigger.AccountingInsertedRoutineTriggerFunctionName,
		saccountingtrigger.AccountingDeletedRoutineTriggerFunctionName,
		saccountingtrigger.AccountingInsertedStationTriggerFunctionName,
		saccountingtrigger.AccountingDeletedStationTriggerFunctionName,
		saccountingtrigger.AccountingMutatedStationTriggerFunctionName,
	} {
		objects = append(objects, spostgres.PermissionObject{
			Type: spostgres.PermissionObjectType_Function,
			Name: functionName,
			Grants: []spostgres.PermissionGrant{
				{
					Runtime:    ctypes.Runtime_Core,
					Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Execute},
				},
			},
		})
		if yjsTriggerFunctions[functionName] {
			objects[len(objects)-1].Grants = append(objects[len(objects)-1].Grants,
				spostgres.PermissionGrant{
					Runtime:    ctypes.Runtime_YjsWorker,
					Privileges: []spostgres.PermissionPrivilege{spostgres.PermissionPrivilege_Execute},
				},
			)
		}
	}
	return objects
}
