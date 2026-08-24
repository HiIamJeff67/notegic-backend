package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UserGender enumcontract.UserGender

func (value *UserGender) ToContractable() *enumcontract.UserGender {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.UserGender(*value)
	return &contractValue
}

func UserGenderToStorable(value *enumcontract.UserGender) *UserGender {
	if value == nil {
		return nil
	}

	storableValue := UserGender(*value)
	return &storableValue
}

const (
	UserGender_Male           UserGender = UserGender(enumcontract.UserGender_Male)
	UserGender_Female         UserGender = UserGender(enumcontract.UserGender_Female)
	UserGender_PreferNotToSay UserGender = UserGender(enumcontract.UserGender_PreferNotToSay)
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

func (g UserGender) Name() string {
	return reflect.TypeOf(g).Name()
}

func (g *UserGender) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*g = UserGender(string(v))
		return nil
	case string:
		*g = UserGender(v)
		return nil
	}
	return scanError(value, g)
}

func (g UserGender) Value() (driver.Value, error) {
	return string(g), nil
}

func (g UserGender) String() string {
	return string(g)
}

func (g *UserGender) IsValidEnum() bool {
	return slices.Contains(AllUserGenders, *g)
}

func ConvertStringToUserGender(enumString string) (*UserGender, error) {
	for _, userGender := range AllUserGenders {
		if string(userGender) == enumString {
			return &userGender, nil
		}
	}
	return nil, fmt.Errorf("invalid user gender: %s", enumString)
}
