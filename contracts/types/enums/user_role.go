package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserRole string

const (
	UserRole_Admin  UserRole = "Admin"
	UserRole_Normal UserRole = "Normal"
	UserRole_Guest  UserRole = "Guest"
)

var AllUserRoles = []UserRole{
	UserRole_Admin,
	UserRole_Normal,
	UserRole_Guest,
}

var AllUserRoleStrings = []string{
	string(UserRole_Admin),
	string(UserRole_Normal),
	string(UserRole_Guest),
}

func (value UserRole) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserRole) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserRole(string(v))
		return nil
	case string:
		*value = UserRole(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserRole) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserRole) String() string {
	return string(value)
}

func (value *UserRole) IsValidEnum() bool {
	return slices.Contains(AllUserRoles, *value)
}

func ConvertStringToUserRole(enumString string) (*UserRole, error) {
	for _, enumValue := range AllUserRoles {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserRole: %s", enumString)
}
