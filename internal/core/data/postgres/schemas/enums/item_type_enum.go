package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type ItemType enumcontract.ItemType

func (value *ItemType) ToContractable() *enumcontract.ItemType {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.ItemType(*value)
	return &contractValue
}

func ItemTypeToStorable(value *enumcontract.ItemType) *ItemType {
	if value == nil {
		return nil
	}

	storableValue := ItemType(*value)
	return &storableValue
}

const (
	ItemType_BlockPack ItemType = ItemType(enumcontract.ItemType_BlockPack)
	ItemType_Material  ItemType = ItemType(enumcontract.ItemType_Material)
)

var AllItemTypes = []ItemType{
	ItemType_BlockPack,
	ItemType_Material,
}

var AllItemTypeStrings = []string{
	string(ItemType_BlockPack),
	string(ItemType_Material),
}

func (it ItemType) Name() string {
	return reflect.TypeOf(it).Name()
}

func (it *ItemType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*it = ItemType(string(v))
		return nil
	case string:
		*it = ItemType(v)
		return nil
	}
	return scanError(value, it)
}

func (it ItemType) Value() (driver.Value, error) {
	return string(it), nil
}

func (it ItemType) String() string {
	return string(it)
}

func (it *ItemType) IsValidEnum() bool {
	return slices.Contains(AllItemTypes, *it)
}

func ConvertStringToItemType(enumString string) (*ItemType, error) {
	for _, itemType := range AllItemTypes {
		if string(itemType) == enumString {
			return &itemType, nil
		}
	}
	return nil, fmt.Errorf("invalid item type: %s", enumString)
}
