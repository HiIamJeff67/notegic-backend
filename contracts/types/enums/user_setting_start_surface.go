package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserSettingStartSurface string

const (
	UserSettingStartSurface_Dashboard UserSettingStartSurface = "Dashboard"
	UserSettingStartSurface_Routines  UserSettingStartSurface = "Routines"
)

var AllUserSettingStartSurfaces = []UserSettingStartSurface{
	UserSettingStartSurface_Dashboard,
	UserSettingStartSurface_Routines,
}

var AllUserSettingStartSurfaceStrings = []string{
	string(UserSettingStartSurface_Dashboard),
	string(UserSettingStartSurface_Routines),
}

func (value UserSettingStartSurface) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserSettingStartSurface) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserSettingStartSurface(string(v))
		return nil
	case string:
		*value = UserSettingStartSurface(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserSettingStartSurface) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserSettingStartSurface) String() string {
	return string(value)
}

func (value *UserSettingStartSurface) IsValidEnum() bool {
	return slices.Contains(AllUserSettingStartSurfaces, *value)
}

func ConvertStringToUserSettingStartSurface(enumString string) (*UserSettingStartSurface, error) {
	for _, enumValue := range AllUserSettingStartSurfaces {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserSettingStartSurface: %s", enumString)
}
