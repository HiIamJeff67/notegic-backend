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

func scanError(value any, enum Enum) error {
	return fmt.Errorf("failed to scan %T into %s", value, enum.Name())
}
