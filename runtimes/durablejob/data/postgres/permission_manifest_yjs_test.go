package postgres

import (
	"slices"
	"testing"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

func TestDatabasePermissionManifestPreservesYjsWorkerAccess(t *testing.T) {
	objects := getDurablePermissionObjects()
	for _, object := range objects {
		if object.Type != spostgres.PermissionObjectType_Database &&
			object.Type != spostgres.PermissionObjectType_Schema {
			continue
		}
		for _, grant := range object.Grants {
			if grant.Runtime != ctypes.Runtime_YjsWorker {
				continue
			}
			if object.Type == spostgres.PermissionObjectType_Database &&
				!slices.Contains(grant.Privileges, spostgres.PermissionPrivilege_Connect) {
				t.Fatalf("DurableJob database manifest must preserve Yjs CONNECT: %v", grant.Privileges)
			}
			if object.Type == spostgres.PermissionObjectType_Schema &&
				!slices.Contains(grant.Privileges, spostgres.PermissionPrivilege_Usage) {
				t.Fatalf("DurableJob schema manifest must preserve Yjs USAGE: %v", grant.Privileges)
			}
		}
	}
}
