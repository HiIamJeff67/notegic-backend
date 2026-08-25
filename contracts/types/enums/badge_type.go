package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type BadgeType string

const (
	BadgeType_Diamond BadgeType = "Diamond"
	BadgeType_Golden  BadgeType = "Golden"
	BadgeType_Silver  BadgeType = "Silver"
	BadgeType_Bronze  BadgeType = "Bronze"
	BadgeType_Steel   BadgeType = "Steel"
)

var AllBadgeTypes = []BadgeType{
	BadgeType_Diamond,
	BadgeType_Golden,
	BadgeType_Silver,
	BadgeType_Bronze,
	BadgeType_Steel,
}

var AllBadgeTypeStrings = []string{
	string(BadgeType_Diamond),
	string(BadgeType_Golden),
	string(BadgeType_Silver),
	string(BadgeType_Bronze),
	string(BadgeType_Steel),
}

func (value BadgeType) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *BadgeType) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = BadgeType(string(v))
		return nil
	case string:
		*value = BadgeType(v)
		return nil
	}
	return scanError(raw, value)
}

func (value BadgeType) Value() (driver.Value, error) {
	return string(value), nil
}

func (value BadgeType) String() string {
	return string(value)
}

func (value *BadgeType) IsValidEnum() bool {
	return slices.Contains(AllBadgeTypes, *value)
}

func ConvertStringToBadgeType(enumString string) (*BadgeType, error) {
	for _, enumValue := range AllBadgeTypes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid BadgeType: %s", enumString)
}
