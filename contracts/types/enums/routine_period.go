package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutinePeriod string

const (
	RoutinePeriod_Daily   RoutinePeriod = "Daily"
	RoutinePeriod_Weekly  RoutinePeriod = "Weekly"
	RoutinePeriod_Monthly RoutinePeriod = "Monthly"
)

var AllRoutinePeriods = []RoutinePeriod{
	RoutinePeriod_Daily,
	RoutinePeriod_Weekly,
	RoutinePeriod_Monthly,
}

var AllRoutinePeriodStrings = []string{
	string(RoutinePeriod_Daily),
	string(RoutinePeriod_Weekly),
	string(RoutinePeriod_Monthly),
}

func (value RoutinePeriod) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutinePeriod) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutinePeriod(string(v))
		return nil
	case string:
		*value = RoutinePeriod(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutinePeriod) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutinePeriod) String() string {
	return string(value)
}

func (value *RoutinePeriod) IsValidEnum() bool {
	return slices.Contains(AllRoutinePeriods, *value)
}

func ConvertStringToRoutinePeriod(enumString string) (*RoutinePeriod, error) {
	for _, enumValue := range AllRoutinePeriods {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutinePeriod: %s", enumString)
}
