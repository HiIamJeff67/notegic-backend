package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type SupportedCurrencyCode string

const (
	SupportedCurrencyCode_USD SupportedCurrencyCode = "USD"
	SupportedCurrencyCode_EUR SupportedCurrencyCode = "EUR"
	SupportedCurrencyCode_JPY SupportedCurrencyCode = "JPY"
	SupportedCurrencyCode_TWD SupportedCurrencyCode = "TWD"
	SupportedCurrencyCode_KRW SupportedCurrencyCode = "KRW"
	SupportedCurrencyCode_CNY SupportedCurrencyCode = "CNY"
)

var AllSupportedCurrencyCodes = []SupportedCurrencyCode{
	SupportedCurrencyCode_USD,
	SupportedCurrencyCode_EUR,
	SupportedCurrencyCode_JPY,
	SupportedCurrencyCode_TWD,
	SupportedCurrencyCode_KRW,
	SupportedCurrencyCode_CNY,
}

var AllSupportedCurrencyCodeStrings = []string{
	string(SupportedCurrencyCode_USD),
	string(SupportedCurrencyCode_EUR),
	string(SupportedCurrencyCode_JPY),
	string(SupportedCurrencyCode_TWD),
	string(SupportedCurrencyCode_KRW),
	string(SupportedCurrencyCode_CNY),
}

func (value SupportedCurrencyCode) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *SupportedCurrencyCode) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = SupportedCurrencyCode(string(v))
		return nil
	case string:
		*value = SupportedCurrencyCode(v)
		return nil
	}
	return scanError(raw, value)
}

func (value SupportedCurrencyCode) Value() (driver.Value, error) {
	return string(value), nil
}

func (value SupportedCurrencyCode) String() string {
	return string(value)
}

func (value *SupportedCurrencyCode) IsValidEnum() bool {
	return slices.Contains(AllSupportedCurrencyCodes, *value)
}

func ConvertStringToSupportedCurrencyCode(enumString string) (*SupportedCurrencyCode, error) {
	for _, enumValue := range AllSupportedCurrencyCodes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid SupportedCurrencyCode: %s", enumString)
}
