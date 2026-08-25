package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type UserGender string

const (
	UserGender_Male           UserGender = "Male"
	UserGender_Female         UserGender = "Female"
	UserGender_PreferNotToSay UserGender = "PreferNotToSay"
)

var AllUserGenders = []UserGender{
	UserGender_Male,
	UserGender_Female,
	UserGender_PreferNotToSay,
}

var AllUserGenderStrings = []string{
	string(UserGender_Male),
	string(UserGender_Female),
	string(UserGender_PreferNotToSay),
}

func (value UserGender) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *UserGender) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = UserGender(string(v))
		return nil
	case string:
		*value = UserGender(v)
		return nil
	}
	return scanError(raw, value)
}

func (value UserGender) Value() (driver.Value, error) {
	return string(value), nil
}

func (value UserGender) String() string {
	return string(value)
}

func (value *UserGender) IsValidEnum() bool {
	return slices.Contains(AllUserGenders, *value)
}

func ConvertStringToUserGender(enumString string) (*UserGender, error) {
	for _, enumValue := range AllUserGenders {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid UserGender: %s", enumString)
}
