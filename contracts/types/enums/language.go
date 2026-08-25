package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type Language string

const (
	Language_English            Language = "English"
	Language_TraditionalChinese Language = "TraditionalChinese"
	Language_SimpleChinese      Language = "SimpleChinese"
	Language_Japanese           Language = "Japanese"
	Language_Korean             Language = "Korean"
)

var AllLanguages = []Language{
	Language_English,
	Language_TraditionalChinese,
	Language_SimpleChinese,
	Language_Japanese,
	Language_Korean,
}

var AllLanguageStrings = []string{
	string(Language_English),
	string(Language_TraditionalChinese),
	string(Language_SimpleChinese),
	string(Language_Japanese),
	string(Language_Korean),
}

func (value Language) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *Language) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = Language(string(v))
		return nil
	case string:
		*value = Language(v)
		return nil
	}
	return scanError(raw, value)
}

func (value Language) Value() (driver.Value, error) {
	return string(value), nil
}

func (value Language) String() string {
	return string(value)
}

func (value *Language) IsValidEnum() bool {
	return slices.Contains(AllLanguages, *value)
}

func ConvertStringToLanguage(enumString string) (*Language, error) {
	for _, enumValue := range AllLanguages {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid Language: %s", enumString)
}
