package enums

import (
	"database/sql/driver"
	"fmt"
)

type Enum interface {
	Name() string
	Scan(value any) error
	Value() (driver.Value, error)
	String() string
	IsValidEnum() bool
}

/* ==================== Temporary Function to Get the Scan Error ==================== */

func scanError(value any, e Enum) error {
	// A Helper Function to Get the Error
	return fmt.Errorf("failed to scan %T into %s", value, e.Name())
}
