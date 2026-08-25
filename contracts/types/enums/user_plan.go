package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserPlan string

const (
	UserPlan_Enterprise UserPlan = "Enterprise"
	UserPlan_Ultimate   UserPlan = "Ultimate"
	UserPlan_Premium    UserPlan = "Premium"
	UserPlan_Pro        UserPlan = "Pro"
	UserPlan_Free       UserPlan = "Free"
)

var AllUserPlans = []UserPlan{
	UserPlan_Enterprise,
	UserPlan_Ultimate,
	UserPlan_Premium,
	UserPlan_Pro,
	UserPlan_Free,
}

var AllUserPlanStrings = []string{
	string(UserPlan_Enterprise),
	string(UserPlan_Ultimate),
	string(UserPlan_Premium),
	string(UserPlan_Pro),
	string(UserPlan_Free),
}

func (value UserPlan) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserPlan) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserPlan(string(v))
		return nil
	case string:
		*value = UserPlan(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserPlan) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserPlan) String() string {
	return string(value)
}

func (value *UserPlan) IsValidEnum() bool {
	return slices.Contains(AllUserPlans, *value)
}

func ConvertStringToUserPlan(enumString string) (*UserPlan, error) {
	for _, enumValue := range AllUserPlans {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserPlan: %s", enumString)
}
