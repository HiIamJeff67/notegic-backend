package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type BadgeType enumcontract.BadgeType

func (value *BadgeType) ToContractable() *enumcontract.BadgeType {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.BadgeType(*value)
	return &contractValue
}

func BadgeTypeToStorable(value *enumcontract.BadgeType) *BadgeType {
	if value == nil {
		return nil
	}

	storableValue := BadgeType(*value)
	return &storableValue
}

const (
	BadgeType_Diamond BadgeType = BadgeType(enumcontract.BadgeType_Diamond)
	BadgeType_Golden  BadgeType = BadgeType(enumcontract.BadgeType_Golden)
	BadgeType_Silver  BadgeType = BadgeType(enumcontract.BadgeType_Silver)
	BadgeType_Bronze  BadgeType = BadgeType(enumcontract.BadgeType_Bronze)
	BadgeType_Steel   BadgeType = BadgeType(enumcontract.BadgeType_Steel)
)

var AllBadgeTypes = []BadgeType{
	BadgeType_Diamond,
	BadgeType_Golden,
	BadgeType_Silver,
	BadgeType_Bronze,
	BadgeType_Steel,
}
var AllBadgeTypeStrings = []string{
	string(BadgeType_Diamond),
	string(BadgeType_Golden),
	string(BadgeType_Silver),
	string(BadgeType_Bronze),
	string(BadgeType_Steel),
}

func (bt BadgeType) Name() string {
	return reflect.TypeOf(bt).Name()
}

func (bt *BadgeType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*bt = BadgeType(string(v))
		return nil
	case string:
		*bt = BadgeType(v)
		return nil
	}
	return scanError(value, bt)
}

func (bt BadgeType) Value() (driver.Value, error) {
	return string(bt), nil
}

func (bt BadgeType) String() string {
	return string(bt)
}

func (bt *BadgeType) IsValidEnum() bool {
	return slices.Contains(AllBadgeTypes, *bt)
}

func ConvertStringToBadgeType(enumString string) (*BadgeType, error) {
	for _, badgeType := range AllBadgeTypes {
		if string(badgeType) == enumString {
			return &badgeType, nil
		}
	}
	return nil, fmt.Errorf("invalid badge type: %s", enumString)
}
