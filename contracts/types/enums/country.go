package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type Country string

const (
	Country_Taiwan                Country = "Taiwan"
	Country_Japan                 Country = "Japan"
	Country_Malaysia              Country = "Malaysia"
	Country_Singapore             Country = "Singapore"
	Country_China                 Country = "China"
	Country_UnitedStatusOfAmerica Country = "UnitedStatesOfAmerica"
	Country_UnitedKingdom         Country = "UnitedKingdom"
	Country_Australia             Country = "Australia"
	Country_Canada                Country = "Canada"
)

var AllCountries = []Country{
	Country_Taiwan,
	Country_Japan,
	Country_Malaysia,
	Country_Singapore,
	Country_China,
	Country_UnitedStatusOfAmerica,
	Country_UnitedKingdom,
	Country_Australia,
	Country_Canada,
}

var AllCountryStrings = []string{
	string(Country_Taiwan),
	string(Country_Japan),
	string(Country_Malaysia),
	string(Country_Singapore),
	string(Country_China),
	string(Country_UnitedStatusOfAmerica),
	string(Country_UnitedKingdom),
	string(Country_Australia),
	string(Country_Canada),
}

func (value Country) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *Country) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = Country(string(v))
		return nil
	case string:
		*value = Country(v)
		return nil
	}
	return scanError(raw, value)
}

func (value Country) Value() (driver.Value, error) {
	return string(value), nil
}

func (value Country) String() string {
	return string(value)
}

func (value *Country) IsValidEnum() bool {
	return slices.Contains(AllCountries, *value)
}

func ConvertStringToCountry(enumString string) (*Country, error) {
	for _, enumValue := range AllCountries {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid Country: %s", enumString)
}
