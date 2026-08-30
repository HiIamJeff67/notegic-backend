package postgres

import (
	"slices"
	"testing"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func TestDatabasePermissionManifestYjsWorkerBoundary(t *testing.T) {
	objects := getCorePermissionObjects()
	tableGrants := make(map[string][]spostgres.PermissionPrivilege)
	functionGrants := make(map[string][]spostgres.PermissionPrivilege)
	var blockTypeGrants []spostgres.PermissionPrivilege

	for _, object := range objects {
		for _, grant := range object.Grants {
			if grant.Runtime != ctypes.Runtime_YjsWorker {
				continue
			}
			switch object.Type {
			case spostgres.PermissionObjectType_Table:
				tableGrants[object.Name] = grant.Privileges
			case spostgres.PermissionObjectType_Function:
				functionGrants[object.Name] = grant.Privileges
			case spostgres.PermissionObjectType_Enum:
				if object.Name == "BlockType" {
					blockTypeGrants = grant.Privileges
				}
			}
		}
	}

	if got := tableGrants[spostgres.TableName_BlockTable.String()]; !samePrivileges(got,
		[]spostgres.PermissionPrivilege{
			spostgres.PermissionPrivilege_Select,
			spostgres.PermissionPrivilege_Insert,
			spostgres.PermissionPrivilege_Update,
			spostgres.PermissionPrivilege_Delete,
		},
	) {
		t.Fatalf("Yjs BlockTable privileges = %v", got)
	}
	for _, table := range []string{
		spostgres.TableName_BlockPackTable.String(),
		spostgres.TableName_BlockPackYjsDocumentTable.String(),
		spostgres.TableName_BlockPackYjsUpdateTable.String(),
	} {
		if _, ok := tableGrants[table]; !ok {
			t.Fatalf("Yjs worker must have a grant on %s", table)
		}
	}
	for _, table := range []string{
		spostgres.TableName_UserTable.String(),
		spostgres.TableName_UserAccountTable.String(),
		spostgres.TableName_PlanLimitationTable.String(),
	} {
		if _, ok := tableGrants[table]; ok {
			t.Fatalf("Yjs worker must not have a grant on %s", table)
		}
	}

	if !samePrivileges(blockTypeGrants, []spostgres.PermissionPrivilege{
		spostgres.PermissionPrivilege_Usage,
	}) {
		t.Fatalf("Yjs BlockType privileges = %v", blockTypeGrants)
	}
	for _, functionName := range []string{
		"trigger_function_accounting_inserted_block",
		"trigger_function_accounting_deleted_block",
	} {
		if !samePrivileges(functionGrants[functionName], []spostgres.PermissionPrivilege{
			spostgres.PermissionPrivilege_Execute,
		}) {
			t.Fatalf("Yjs %s privileges = %v", functionName, functionGrants[functionName])
		}
	}
}

func samePrivileges(left, right []spostgres.PermissionPrivilege) bool {
	if len(left) != len(right) {
		return false
	}
	for _, privilege := range left {
		if !slices.Contains(right, privilege) {
			return false
		}
	}
	return true
}
