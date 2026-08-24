package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UserPlan enumcontract.UserPlan

func (value *UserPlan) ToContractable() *enumcontract.UserPlan {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UserPlan(*value)
	return &contractValue
}

func UserPlanToStorable(value *enumcontract.UserPlan) *UserPlan {
	if value == nil {
		return nil
	}

	storableValue := UserPlan(*value)
	return &storableValue
}

const (
	UserPlan_Enterprise UserPlan = UserPlan(enumcontract.UserPlan_Enterprise)
	UserPlan_Ultimate   UserPlan = UserPlan(enumcontract.UserPlan_Ultimate)
	UserPlan_Premium    UserPlan = UserPlan(enumcontract.UserPlan_Premium)
	UserPlan_Pro        UserPlan = UserPlan(enumcontract.UserPlan_Pro)
	UserPlan_Free       UserPlan = UserPlan(enumcontract.UserPlan_Free)
)

// All the userPlans placing in the descending order
var AllUserPlans = []UserPlan{
	UserPlan_Enterprise,
	UserPlan_Ultimate,
	UserPlan_Premium,
	UserPlan_Pro,
	UserPlan_Free,
}

// All the userPlan strings placing in the descending order
var AllUserPlanStrings = []string{
	string(UserPlan_Enterprise),
	string(UserPlan_Ultimate),
	string(UserPlan_Premium),
	string(UserPlan_Pro),
	string(UserPlan_Free),
}

func (p UserPlan) Name() string {
	return reflect.TypeOf(p).Name()
}

func (p *UserPlan) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*p = UserPlan(string(v))
		return nil
	case string:
		*p = UserPlan(v)
		return nil
	}
	return scanError(value, p)
}

func (p UserPlan) Value() (driver.Value, error) {
	return string(p), nil
}

func (p UserPlan) String() string {
	return string(p)
}

func (p *UserPlan) IsValidEnum() bool {
	return slices.Contains(AllUserPlans, *p)
}

func ConvertStringToUserPlan(enumString string) (*UserPlan, error) {
	for _, userPlan := range AllUserPlans {
		if string(userPlan) == enumString {
			return &userPlan, nil
		}
	}
	return nil, fmt.Errorf("invalid user plan: %s", enumString)
}
