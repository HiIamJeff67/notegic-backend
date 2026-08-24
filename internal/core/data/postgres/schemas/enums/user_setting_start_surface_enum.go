package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UserSettingStartSurface enumcontract.UserSettingStartSurface

const (
	UserSettingStartSurface_Dashboard UserSettingStartSurface = UserSettingStartSurface(enumcontract.UserSettingStartSurface_Dashboard)
	UserSettingStartSurface_Routines  UserSettingStartSurface = UserSettingStartSurface(enumcontract.UserSettingStartSurface_Routines)
)

var AllUserSettingStartSurfaces = []UserSettingStartSurface{
	UserSettingStartSurface_Dashboard,
	UserSettingStartSurface_Routines,
}

var AllUserSettingStartSurfaceStrings = []string{
	string(UserSettingStartSurface_Dashboard),
	string(UserSettingStartSurface_Routines),
}

func (s UserSettingStartSurface) Name() string { return reflect.TypeOf(s).Name() }

func (value *UserSettingStartSurface) ToContractable() *enumcontract.UserSettingStartSurface {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UserSettingStartSurface(*value)
	return &contractValue
}

func UserSettingStartSurfaceToStorable(value *enumcontract.UserSettingStartSurface) *UserSettingStartSurface {
	if value == nil {
		return nil
	}

	storableValue := UserSettingStartSurface(*value)
	return &storableValue
}

func (s *UserSettingStartSurface) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*s = UserSettingStartSurface(string(v))
		return nil
	case string:
		*s = UserSettingStartSurface(v)
		return nil
	}
	return scanError(value, s)
}

func (s UserSettingStartSurface) Value() (driver.Value, error) { return string(s), nil }
func (s UserSettingStartSurface) String() string               { return string(s) }
func (s *UserSettingStartSurface) IsValidEnum() bool {
	return slices.Contains(AllUserSettingStartSurfaces, *s)
}

func ConvertStringToUserSettingStartSurface(value string) (*UserSettingStartSurface, error) {
	for _, surface := range AllUserSettingStartSurfaces {
		if string(surface) == value {
			return &surface, nil
		}
	}
	return nil, fmt.Errorf("invalid user setting start surface: %s", value)
}
