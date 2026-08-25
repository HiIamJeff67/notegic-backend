package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserSettingDensity string

const (
	UserSettingDensity_Comfortable UserSettingDensity = "Comfortable"
	UserSettingDensity_Balanced    UserSettingDensity = "Balanced"
	UserSettingDensity_Compact     UserSettingDensity = "Compact"
)

var AllUserSettingDensities = []UserSettingDensity{
	UserSettingDensity_Comfortable,
	UserSettingDensity_Balanced,
	UserSettingDensity_Compact,
}

var AllUserSettingDensityStrings = []string{
	string(UserSettingDensity_Comfortable),
	string(UserSettingDensity_Balanced),
	string(UserSettingDensity_Compact),
}

func (value UserSettingDensity) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserSettingDensity) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserSettingDensity(string(v))
		return nil
	case string:
		*value = UserSettingDensity(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserSettingDensity) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserSettingDensity) String() string {
	return string(value)
}

func (value *UserSettingDensity) IsValidEnum() bool {
	return slices.Contains(AllUserSettingDensities, *value)
}

func ConvertStringToUserSettingDensity(enumString string) (*UserSettingDensity, error) {
	for _, enumValue := range AllUserSettingDensities {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserSettingDensity: %s", enumString)
}
