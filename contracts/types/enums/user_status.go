package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserStatus string

const (
	UserStatus_Online       UserStatus = "Online"
	UserStatus_AFK          UserStatus = "AFK"
	UserStatus_DoNotDisturb UserStatus = "DoNotDisturb"
	UserStatus_Offline      UserStatus = "Offline"
)

var AllUserStatuses = []UserStatus{
	UserStatus_Online,
	UserStatus_AFK,
	UserStatus_DoNotDisturb,
	UserStatus_Offline,
}

var AllUserStatusStrings = []string{
	string(UserStatus_Online),
	string(UserStatus_AFK),
	string(UserStatus_DoNotDisturb),
	string(UserStatus_Offline),
}

func (value UserStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserStatus(string(v))
		return nil
	case string:
		*value = UserStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserStatus) String() string {
	return string(value)
}

func (value *UserStatus) IsValidEnum() bool {
	return slices.Contains(AllUserStatuses, *value)
}

func ConvertStringToUserStatus(enumString string) (*UserStatus, error) {
	for _, enumValue := range AllUserStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserStatus: %s", enumString)
}
