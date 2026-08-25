package postgres

import ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

// MigrationManifest contains the database objects a runtime is responsible
// for migrating. It intentionally does not contain database permissions.
type MigrationManifest struct {
	Runtime     ctypes.Runtime
	Enums       map[string][]string
	Tables      []any
	Views       []string
	Triggers    []string
	Constraints []string
}

func (manifest MigrationManifest) IsFor(runtime ctypes.Runtime) bool {
	return manifest.Runtime == runtime
}
