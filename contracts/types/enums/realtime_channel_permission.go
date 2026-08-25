package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type ChannelPermission string

const (
	ChannelPermission_Read  ChannelPermission = "read"
	ChannelPermission_Write ChannelPermission = "write"
)

func (p ChannelPermission) AllowedAccessControlPermissions() []AccessControlPermission {
	switch p {
	case ChannelPermission_Read:
		return []AccessControlPermission{
			AccessControlPermission_Owner,
			AccessControlPermission_Admin,
			AccessControlPermission_Write,
			AccessControlPermission_Read,
		}
	case ChannelPermission_Write:
		return []AccessControlPermission{
			AccessControlPermission_Owner,
			AccessControlPermission_Admin,
			AccessControlPermission_Write,
		}
	default:
		return nil
	}
}

var AllChannelPermissions = []ChannelPermission{
	ChannelPermission_Read,
	ChannelPermission_Write,
}

var AllChannelPermissionsStrings = []string{
	string(ChannelPermission_Read),
	string(ChannelPermission_Write),
}

func (value ChannelPermission) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *ChannelPermission) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = ChannelPermission(string(v))
		return nil
	case string:
		*value = ChannelPermission(v)
		return nil
	}
	return scanError(raw, value)
}

func (value ChannelPermission) Value() (driver.Value, error) {
	return string(value), nil
}

func (value ChannelPermission) String() string {
	return string(value)
}

func (value *ChannelPermission) IsValidEnum() bool {
	return slices.Contains(AllChannelPermissions, *value)
}

func ConvertStringToConvertStringToChannelPermission(enumString string) (*ChannelPermission, error) {
	for _, enumValue := range AllChannelPermissions {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid ChannelPermission: %s", enumString)
}
