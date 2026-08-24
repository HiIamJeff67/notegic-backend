package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type CountryCode enumcontract.CountryCode

func (value *CountryCode) ToContractable() *enumcontract.CountryCode {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.CountryCode(*value)
	return &contractValue
}

func CountryCodeToStorable(value *enumcontract.CountryCode) *CountryCode {
	if value == nil {
		return nil
	}

	storableValue := CountryCode(*value)
	return &storableValue
}

const (
	CountryCode_Taiwan        CountryCode = CountryCode(enumcontract.CountryCode_Taiwan)
	CountryCode_Japan         CountryCode = CountryCode(enumcontract.CountryCode_Japan)
	CountryCode_Malaysia      CountryCode = CountryCode(enumcontract.CountryCode_Malaysia)
	CountryCode_Singapore     CountryCode = CountryCode(enumcontract.CountryCode_Singapore)
	CountryCode_China         CountryCode = CountryCode(enumcontract.CountryCode_China)
	CountryCode_NANP          CountryCode = CountryCode(enumcontract.CountryCode_NANP)
	CountryCode_UnitedKingdom CountryCode = CountryCode(enumcontract.CountryCode_UnitedKingdom)
	CountryCode_Australia     CountryCode = CountryCode(enumcontract.CountryCode_Australia)
)

var AllCountryCodes = []CountryCode{
	CountryCode_Taiwan,
	CountryCode_Japan,
	CountryCode_Malaysia,
	CountryCode_Singapore,
	CountryCode_China,
	CountryCode_NANP, // NANP stands for North American Numbering Plan, it's used in United States of America and Canada
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

func (cc CountryCode) Name() string {
	return reflect.TypeOf(cc).Name()
}

func (cc *CountryCode) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*cc = CountryCode(string(v))
		return nil
	case string:
		*cc = CountryCode(v)
		return nil
	}
	return scanError(value, cc)
}

func (cc CountryCode) Value() (driver.Value, error) {
	return string(cc), nil
}

func (cc CountryCode) String() string {
	return string(cc)
}

func (cc *CountryCode) IsValidEnum() bool {
	return slices.Contains(AllCountryCodes, *cc)
}

func ConvertStringToCountryCode(enumString string) (*CountryCode, error) {
	for _, countryCode := range AllCountryCodes {
		if string(countryCode) == enumString {
			return &countryCode, nil
		}
	}
	return nil, fmt.Errorf("invalid country code: %s", enumString)
}
