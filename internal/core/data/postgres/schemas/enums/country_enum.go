package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type Country enumcontract.Country

func (value *Country) ToContractable() *enumcontract.Country {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.Country(*value)
	return &contractValue
}

func CountryToStorable(value *enumcontract.Country) *Country {
	if value == nil {
		return nil
	}

	storableValue := Country(*value)
	return &storableValue
}

const (
	Country_Taiwan                Country = Country(enumcontract.Country_Taiwan)
	Country_Japan                 Country = Country(enumcontract.Country_Japan)
	Country_Malaysia              Country = Country(enumcontract.Country_Malaysia)
	Country_Singapore             Country = Country(enumcontract.Country_Singapore)
	Country_China                 Country = Country(enumcontract.Country_China)
	Country_UnitedStatusOfAmerica Country = Country(enumcontract.Country_UnitedStatusOfAmerica)
	Country_UnitedKingdom         Country = Country(enumcontract.Country_UnitedKingdom)
	Country_Australia             Country = Country(enumcontract.Country_Australia)
	Country_Canada                Country = Country(enumcontract.Country_Canada)
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

func (c Country) Name() string {
	return reflect.TypeOf(c).Name()
}

func (c *Country) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*c = Country(string(v))
		return nil
	case string:
		*c = Country(v)
		return nil
	}
	return scanError(value, c)
}

func (c Country) Value() (driver.Value, error) {
	return string(c), nil
}

func (c Country) String() string {
	return string(c)
}

func (c *Country) IsValidEnum() bool {
	return slices.Contains(AllCountries, *c)
}

func ConvertStringToCountry(enumString string) (*Country, error) {
	for _, country := range AllCountries {
		if string(country) == enumString {
			return &country, nil
		}
	}
	return nil, fmt.Errorf("invalid country: %s", enumString)
}
