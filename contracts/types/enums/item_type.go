package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type ItemType string

const (
	ItemType_BlockPack ItemType = "BlockPack"
	ItemType_Material  ItemType = "Material"
)

var AllItemTypes = []ItemType{
	ItemType_BlockPack,
	ItemType_Material,
}

var AllItemTypeStrings = []string{
	string(ItemType_BlockPack),
	string(ItemType_Material),
}

func (value ItemType) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *ItemType) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = ItemType(string(v))
		return nil
	case string:
		*value = ItemType(v)
		return nil
	}
	return scanError(raw, value)
}

func (value ItemType) Value() (driver.Value, error) {
	return string(value), nil
}

func (value ItemType) String() string {
	return string(value)
}

func (value *ItemType) IsValidEnum() bool {
	return slices.Contains(AllItemTypes, *value)
}

func ConvertStringToItemType(enumString string) (*ItemType, error) {
	for _, enumValue := range AllItemTypes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid ItemType: %s", enumString)
}
