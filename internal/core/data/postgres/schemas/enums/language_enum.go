package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type Language enumcontract.Language

func (value *Language) ToContractable() *enumcontract.Language {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.Language(*value)
	return &contractValue
}

func LanguageToStorable(value *enumcontract.Language) *Language {
	if value == nil {
		return nil
	}

	storableValue := Language(*value)
	return &storableValue
}

const (
	Language_English            Language = Language(enumcontract.Language_English)
	Language_TraditionalChinese Language = Language(enumcontract.Language_TraditionalChinese)
	Language_SimpleChinese      Language = Language(enumcontract.Language_SimpleChinese)
	Language_Japanese           Language = Language(enumcontract.Language_Japanese)
	Language_Korean             Language = Language(enumcontract.Language_Korean)
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

func (l Language) Name() string {
	return reflect.TypeOf(l).Name()
}

func (l *Language) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*l = Language(string(v))
		return nil
	case string:
		*l = Language(v)
		return nil
	}
	return scanError(value, l)
}

func (l Language) Value() (driver.Value, error) {
	return string(l), nil
}

func (l Language) String() string {
	return string(l)
}

func (l *Language) IsValidEnum() bool {
	return slices.Contains(AllLanguages, *l)
}

func ConvertStringToLanguage(enumString string) (*Language, error) {
	for _, language := range AllLanguages {
		if string(language) == enumString {
			return &language, nil
		}
	}
	return nil, fmt.Errorf("invalid language: %s", enumString)
}
