package postgres

import (
	"testing"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func TestDatabasePermissionManifestRoutineTaskRuntimeBoundary(t *testing.T) {
	objects := getCorePermissionObjects()
	for _, object := range objects {
		if object.Type != spostgres.PermissionObjectType_Table || object.Name != spostgres.TableName_RoutineTable.String() {
			continue
		}
		for _, grant := range object.Grants {
			if grant.Runtime == ctypes.Runtime_DurableJob && !samePrivileges(grant.Privileges, []spostgres.PermissionPrivilege{
				spostgres.PermissionPrivilege_Select,
				spostgres.PermissionPrivilege_Insert,
				spostgres.PermissionPrivilege_Update,
				spostgres.PermissionPrivilege_Delete,
			}) {
				t.Fatalf("DurableJob RoutineTable privileges = %v", grant.Privileges)
			}
		}
		return
	}

	t.Fatal("RoutineTable must grant DurableJob routine automation access")
}
