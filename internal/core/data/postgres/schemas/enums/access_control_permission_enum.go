package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type AccessControlPermission enumcontract.AccessControlPermission

func (value *AccessControlPermission) ToContractable() *enumcontract.AccessControlPermission {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.AccessControlPermission(*value)
	return &contractValue
}

func AccessControlPermissionToStorable(value *enumcontract.AccessControlPermission) *AccessControlPermission {
	if value == nil {
		return nil
	}

	storableValue := AccessControlPermission(*value)
	return &storableValue
}

const (
	AccessControlPermission_Read  AccessControlPermission = AccessControlPermission(enumcontract.AccessControlPermission_Read)
	AccessControlPermission_Write AccessControlPermission = AccessControlPermission(enumcontract.AccessControlPermission_Write)
	AccessControlPermission_Admin AccessControlPermission = AccessControlPermission(enumcontract.AccessControlPermission_Admin)
	AccessControlPermission_Owner AccessControlPermission = AccessControlPermission(enumcontract.AccessControlPermission_Owner)
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

func (a AccessControlPermission) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a *AccessControlPermission) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*a = AccessControlPermission(string(v))
		return nil
	case string:
		*a = AccessControlPermission(v)
		return nil
	}
	return scanError(value, a)
}

func (a AccessControlPermission) Value() (driver.Value, error) {
	return string(a), nil
}

func (a AccessControlPermission) String() string {
	return string(a)
}

func (a *AccessControlPermission) IsValidEnum() bool {
	return slices.Contains(AllAccessControlPermissions, *a)
}

func ConvertStringToAccessControlPermission(enumString string) (*AccessControlPermission, error) {
	for _, accessControlPermission := range AllAccessControlPermissions {
		if string(accessControlPermission) == enumString {
			return &accessControlPermission, nil
		}
	}
	return nil, fmt.Errorf("invalid access control permission: %s", enumString)
}
