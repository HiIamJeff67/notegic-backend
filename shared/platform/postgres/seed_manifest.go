package postgres

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
)

// SeedManifest contains the seed statements owned by one runtime. Seed data
// has no independent PostgreSQL permission object; the runtime role's table
// privileges determine whether these statements can write successfully.
type SeedManifest struct {
	Runtime ctypes.Runtime
	SQLs    []string
}

func (manifest SeedManifest) IsFor(runtime ctypes.Runtime) bool {
	return manifest.Runtime == runtime
}

// Seed executes a runtime's seed manifest through an admin connection while
// enforcing the runtime role for every statement in the transaction.
func Seed(db *gorm.DB, runtime ctypes.Runtime, manifest SeedManifest) error {
	if db == nil {
		return errors.New("admin database connection is required")
	}
	if !runtime.IsValid() {
		return fmt.Errorf("invalid runtime %q", runtime)
	}
	if !manifest.IsFor(runtime) {
		return fmt.Errorf("runtime %q cannot seed manifest for runtime %q", runtime, manifest.Runtime)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL ROLE " + quoteIdentifier(runtime.RoleName())).Error; err != nil {
			return err
		}
		return migrateSQL(tx, manifest.SQLs, "seed data", false)
	})
}
