package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type AccessControlPermission string

const (
	AccessControlPermission_Read  AccessControlPermission = "Read"
	AccessControlPermission_Write AccessControlPermission = "Write"
	AccessControlPermission_Admin AccessControlPermission = "Admin"
	AccessControlPermission_Owner AccessControlPermission = "Owner"
)

var AllAccessControlPermissions = []AccessControlPermission{
	AccessControlPermission_Read,
	AccessControlPermission_Write,
	AccessControlPermission_Admin,
	AccessControlPermission_Owner,
}

var AllAccessControlPermissionStrings = []string{
	string(AccessControlPermission_Read),
	string(AccessControlPermission_Write),
	string(AccessControlPermission_Admin),
	string(AccessControlPermission_Owner),
}

func (value AccessControlPermission) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *AccessControlPermission) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = AccessControlPermission(string(v))
		return nil
	case string:
		*value = AccessControlPermission(v)
		return nil
	}
	return scanError(raw, value)
}

func (value AccessControlPermission) Value() (driver.Value, error) {
	return string(value), nil
}

func (value AccessControlPermission) String() string {
	return string(value)
}

func (value *AccessControlPermission) IsValidEnum() bool {
	return slices.Contains(AllAccessControlPermissions, *value)
}

func ConvertStringToAccessControlPermission(enumString string) (*AccessControlPermission, error) {
	for _, enumValue := range AllAccessControlPermissions {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid AccessControlPermission: %s", enumString)
}
