package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type BillingIntervalUnit string

const (
	BillingIntervalUnit_Day   BillingIntervalUnit = "DAY"
	BillingIntervalUnit_Week  BillingIntervalUnit = "WEEK"
	BillingIntervalUnit_Month BillingIntervalUnit = "MONTH"
	BillingIntervalUnit_Year  BillingIntervalUnit = "YEAR"
)

var AllBillingIntervalUnits = []BillingIntervalUnit{
	BillingIntervalUnit_Day,
	BillingIntervalUnit_Week,
	BillingIntervalUnit_Month,
	BillingIntervalUnit_Year,
}

var AllBillingIntervalUnitStrings = []string{
	string(BillingIntervalUnit_Day),
	string(BillingIntervalUnit_Week),
	string(BillingIntervalUnit_Month),
	string(BillingIntervalUnit_Year),
}

func (value BillingIntervalUnit) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *BillingIntervalUnit) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = BillingIntervalUnit(string(v))
		return nil
	case string:
		*value = BillingIntervalUnit(v)
		return nil
	}
	return scanError(raw, value)
}

func (value BillingIntervalUnit) Value() (driver.Value, error) {
	return string(value), nil
}

func (value BillingIntervalUnit) String() string {
	return string(value)
}

func (value *BillingIntervalUnit) IsValidEnum() bool {
	return slices.Contains(AllBillingIntervalUnits, *value)
}

func ConvertStringToBillingIntervalUnit(enumString string) (*BillingIntervalUnit, error) {
	for _, enumValue := range AllBillingIntervalUnits {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid BillingIntervalUnit: %s", enumString)
}
