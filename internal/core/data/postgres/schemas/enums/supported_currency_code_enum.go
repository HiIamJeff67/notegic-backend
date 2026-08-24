package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type SupportedCurrencyCode enumcontract.SupportedCurrencyCode

func (value *SupportedCurrencyCode) ToContractable() *enumcontract.SupportedCurrencyCode {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.SupportedCurrencyCode(*value)
	return &contractValue
}

func SupportedCurrencyCodeToStorable(value *enumcontract.SupportedCurrencyCode) *SupportedCurrencyCode {
	if value == nil {
		return nil
	}

	storableValue := SupportedCurrencyCode(*value)
	return &storableValue
}

const (
	SupportedCurrencyCode_USD SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_USD)
	SupportedCurrencyCode_EUR SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_EUR)
	SupportedCurrencyCode_JPY SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_JPY)
	SupportedCurrencyCode_TWD SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_TWD)
	SupportedCurrencyCode_KRW SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_KRW)
	SupportedCurrencyCode_CNY SupportedCurrencyCode = SupportedCurrencyCode(enumcontract.SupportedCurrencyCode_CNY)
)

/* ============================== All instances ============================== */

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

func (scc SupportedCurrencyCode) Name() string {
	return reflect.TypeOf(scc).Name()
}

func (scc *SupportedCurrencyCode) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*scc = SupportedCurrencyCode(string(v))
		return nil
	case string:
		*scc = SupportedCurrencyCode(v)
		return nil
	}
	return scanError(value, scc)
}

func (scc SupportedCurrencyCode) Value() (driver.Value, error) {
	return string(scc), nil
}

func (scc SupportedCurrencyCode) String() string {
	return string(scc)
}

func (scc *SupportedCurrencyCode) IsValidEnum() bool {
	return slices.Contains(AllSupportedCurrencyCodes, *scc)
}

func ConvertStringToSupportedCurrencyCode(enumString string) (*SupportedCurrencyCode, error) {
	for _, supportedCurrencyCode := range AllSupportedCurrencyCodes {
		if string(supportedCurrencyCode) == enumString {
			return &supportedCurrencyCode, nil
		}
	}
	return nil, fmt.Errorf("invalid supported currency code: %s", enumString)
}
