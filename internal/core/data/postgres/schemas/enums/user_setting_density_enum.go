package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UserSettingDensity enumcontract.UserSettingDensity

const (
	UserSettingDensity_Comfortable UserSettingDensity = UserSettingDensity(enumcontract.UserSettingDensity_Comfortable)
	UserSettingDensity_Balanced    UserSettingDensity = UserSettingDensity(enumcontract.UserSettingDensity_Balanced)
	UserSettingDensity_Compact     UserSettingDensity = UserSettingDensity(enumcontract.UserSettingDensity_Compact)
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

func (s UserSettingDensity) Name() string { return reflect.TypeOf(s).Name() }

func (value *UserSettingDensity) ToContractable() *enumcontract.UserSettingDensity {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UserSettingDensity(*value)
	return &contractValue
}

func UserSettingDensityToStorable(value *enumcontract.UserSettingDensity) *UserSettingDensity {
	if value == nil {
		return nil
	}

	storableValue := UserSettingDensity(*value)
	return &storableValue
}

func (s *UserSettingDensity) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*s = UserSettingDensity(string(v))
		return nil
	case string:
		*s = UserSettingDensity(v)
		return nil
	}
	return scanError(value, s)
}

func (s UserSettingDensity) Value() (driver.Value, error) { return string(s), nil }
func (s UserSettingDensity) String() string               { return string(s) }
func (s *UserSettingDensity) IsValidEnum() bool           { return slices.Contains(AllUserSettingDensities, *s) }

func ConvertStringToUserSettingDensity(value string) (*UserSettingDensity, error) {
	for _, density := range AllUserSettingDensities {
		if string(density) == value {
			return &density, nil
		}
	}
	return nil, fmt.Errorf("invalid user setting density: %s", value)
}
