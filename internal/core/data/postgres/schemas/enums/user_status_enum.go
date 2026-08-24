package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UserStatus enumcontract.UserStatus

func (value *UserStatus) ToContractable() *enumcontract.UserStatus {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UserStatus(*value)
	return &contractValue
}

func UserStatusToStorable(value *enumcontract.UserStatus) *UserStatus {
	if value == nil {
		return nil
	}

	storableValue := UserStatus(*value)
	return &storableValue
}

const (
	UserStatus_Online       UserStatus = UserStatus(enumcontract.UserStatus_Online)
	UserStatus_AFK          UserStatus = UserStatus(enumcontract.UserStatus_AFK)
	UserStatus_DoNotDisturb UserStatus = UserStatus(enumcontract.UserStatus_DoNotDisturb)
	UserStatus_Offline      UserStatus = UserStatus(enumcontract.UserStatus_Offline)
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

func (s UserStatus) Name() string {
	return reflect.TypeOf(s).Name()
}

func (s *UserStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*s = UserStatus(string(v))
		return nil
	case string:
		*s = UserStatus(v)
		return nil
	}
	return scanError(value, s)
}

func (s UserStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s UserStatus) String() string {
	return string(s)
}

func (s *UserStatus) IsValidEnum() bool {
	return slices.Contains(AllUserStatuses, *s)
}

func ConvertStringToUserStatus(enumString string) (*UserStatus, error) {
	for _, userStatus := range AllUserStatuses {
		if string(userStatus) == enumString {
			return &userStatus, nil
		}
	}
	return nil, fmt.Errorf("invalid user status: %s", enumString)
}
