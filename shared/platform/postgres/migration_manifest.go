package postgres

import types "github.com/HiIamJeff67/notegic-backend/contracts/types"

// MigrationManifest contains the database objects a runtime is responsible
// for migrating. It intentionally does not contain database permissions.
type MigrationManifest struct {
	Runtime     types.Runtime
	Enums       map[string][]string
	Tables      []any
	Views       []string
	Triggers    []string
	Constraints []string
}

func (manifest MigrationManifest) IsFor(runtime types.Runtime) bool {
	return manifest.Runtime == runtime
}
