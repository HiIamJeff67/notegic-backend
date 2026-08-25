package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type CountryCode string

const (
	CountryCode_Taiwan        CountryCode = "+886"
	CountryCode_Japan         CountryCode = "+81"
	CountryCode_Malaysia      CountryCode = "+60"
	CountryCode_Singapore     CountryCode = "+65"
	CountryCode_China         CountryCode = "+86"
	CountryCode_NANP          CountryCode = "+1"
	CountryCode_UnitedKingdom CountryCode = "+44"
	CountryCode_Australia     CountryCode = "+61"
)

var AllCountryCodes = []CountryCode{
	CountryCode_Taiwan,
	CountryCode_Japan,
	CountryCode_Malaysia,
	CountryCode_Singapore,
	CountryCode_China,
	CountryCode_NANP,
	CountryCode_UnitedKingdom,
	CountryCode_Australia,
}

var AllCountryCodeStrings = []string{
	string(CountryCode_Taiwan),
	string(CountryCode_Japan),
	string(CountryCode_Malaysia),
	string(CountryCode_Singapore),
	string(CountryCode_China),
	string(CountryCode_NANP),
	string(CountryCode_UnitedKingdom),
	string(CountryCode_Australia),
}

func (value CountryCode) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *CountryCode) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = CountryCode(string(v))
		return nil
	case string:
		*value = CountryCode(v)
		return nil
	}
	return scanError(raw, value)
}

func (value CountryCode) Value() (driver.Value, error) {
	return string(value), nil
}

func (value CountryCode) String() string {
	return string(value)
}

func (value *CountryCode) IsValidEnum() bool {
	return slices.Contains(AllCountryCodes, *value)
}

func ConvertStringToCountryCode(enumString string) (*CountryCode, error) {
	for _, enumValue := range AllCountryCodes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid CountryCode: %s", enumString)
}
