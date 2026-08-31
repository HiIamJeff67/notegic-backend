package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineTaskPurpose string

type RoutineTaskObjectKind string

const (
	RoutineTaskObjectKind_Container RoutineTaskObjectKind = "Container"
	RoutineTaskObjectKind_Core      RoutineTaskObjectKind = "Core"
)

const (
	// Current automation purposes. RootShelf, Station, and individual Block are
	// intentionally not automation targets.
	RoutineTaskPurpose_GetSubShelf     RoutineTaskPurpose = "GetSubShelf"
	RoutineTaskPurpose_CreateSubShelf  RoutineTaskPurpose = "CreateSubShelf" // create a sub shelf with nothing inside of it
	RoutineTaskPurpose_UpdateSubShelf  RoutineTaskPurpose = "UpdateSubShelf" // update the columns of the given sub shelf
	RoutineTaskPurpose_DeleteSubShelf  RoutineTaskPurpose = "DeleteSubShelf"
	RoutineTaskPurpose_GetBlockPack    RoutineTaskPurpose = "GetBlockPack"
	RoutineTaskPurpose_CreateBlockPack RoutineTaskPurpose = "CreateBlockPack" // create a block pack with the given content within the routine task payload
	RoutineTaskPurpose_UpdateBlockPack RoutineTaskPurpose = "UpdateBlockPack" // update blocks in the block pack
	RoutineTaskPurpose_DeleteBlockPack RoutineTaskPurpose = "DeleteBlockPack"
	RoutineTaskPurpose_GetRoutine      RoutineTaskPurpose = "GetRoutine"
	RoutineTaskPurpose_CreateRoutine   RoutineTaskPurpose = "CreateRoutine" // create a routine with no links
	RoutineTaskPurpose_UpdateRoutine   RoutineTaskPurpose = "UpdateRoutine" // update the columns of the given routine, excluded links to it
	RoutineTaskPurpose_DeleteRoutine   RoutineTaskPurpose = "DeleteRoutine"
	RoutineTaskPurpose_GetMaterial     RoutineTaskPurpose = "GetMaterial"
	RoutineTaskPurpose_CreateMaterial  RoutineTaskPurpose = "CreateMaterial"
	RoutineTaskPurpose_UpdateMaterial  RoutineTaskPurpose = "UpdateMaterial"
	RoutineTaskPurpose_DeleteMaterial  RoutineTaskPurpose = "DeleteMaterial"
)

var AllRoutineTaskPurposes = []RoutineTaskPurpose{
	RoutineTaskPurpose_GetSubShelf,
	RoutineTaskPurpose_CreateSubShelf,
	RoutineTaskPurpose_UpdateSubShelf,
	RoutineTaskPurpose_DeleteSubShelf,
	RoutineTaskPurpose_GetBlockPack,
	RoutineTaskPurpose_CreateBlockPack,
	RoutineTaskPurpose_UpdateBlockPack,
	RoutineTaskPurpose_DeleteBlockPack,
	RoutineTaskPurpose_GetRoutine,
	RoutineTaskPurpose_CreateRoutine,
	RoutineTaskPurpose_UpdateRoutine,
	RoutineTaskPurpose_DeleteRoutine,
	RoutineTaskPurpose_GetMaterial,
	RoutineTaskPurpose_CreateMaterial,
	RoutineTaskPurpose_UpdateMaterial,
	RoutineTaskPurpose_DeleteMaterial,
}

var AllRoutineTaskPurposeStrings = []string{
	string(RoutineTaskPurpose_GetSubShelf),
	string(RoutineTaskPurpose_CreateSubShelf),
	string(RoutineTaskPurpose_UpdateSubShelf),
	string(RoutineTaskPurpose_DeleteSubShelf),
	string(RoutineTaskPurpose_GetBlockPack),
	string(RoutineTaskPurpose_CreateBlockPack),
	string(RoutineTaskPurpose_UpdateBlockPack),
	string(RoutineTaskPurpose_DeleteBlockPack),
	string(RoutineTaskPurpose_GetRoutine),
	string(RoutineTaskPurpose_CreateRoutine),
	string(RoutineTaskPurpose_UpdateRoutine),
	string(RoutineTaskPurpose_DeleteRoutine),
	string(RoutineTaskPurpose_GetMaterial),
	string(RoutineTaskPurpose_CreateMaterial),
	string(RoutineTaskPurpose_UpdateMaterial),
	string(RoutineTaskPurpose_DeleteMaterial),
}

func (value RoutineTaskPurpose) ObjectKind() (RoutineTaskObjectKind, bool) {
	switch value {
	case RoutineTaskPurpose_GetSubShelf,
		RoutineTaskPurpose_CreateSubShelf,
		RoutineTaskPurpose_UpdateSubShelf,
		RoutineTaskPurpose_DeleteSubShelf,
		RoutineTaskPurpose_GetRoutine,
		RoutineTaskPurpose_CreateRoutine,
		RoutineTaskPurpose_UpdateRoutine,
		RoutineTaskPurpose_DeleteRoutine:
		return RoutineTaskObjectKind_Container, true
	case RoutineTaskPurpose_GetBlockPack,
		RoutineTaskPurpose_CreateBlockPack,
		RoutineTaskPurpose_UpdateBlockPack,
		RoutineTaskPurpose_DeleteBlockPack,
		RoutineTaskPurpose_GetMaterial,
		RoutineTaskPurpose_CreateMaterial,
		RoutineTaskPurpose_UpdateMaterial,
		RoutineTaskPurpose_DeleteMaterial:
		return RoutineTaskObjectKind_Core, true
	default:
		return "", false
	}
}

func (value RoutineTaskPurpose) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineTaskPurpose) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineTaskPurpose(string(v))
		return nil
	case string:
		*value = RoutineTaskPurpose(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineTaskPurpose) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineTaskPurpose) String() string {
	return string(value)
}

func (value *RoutineTaskPurpose) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskPurposes, *value)
}

func ConvertStringToRoutineTaskPurpose(enumString string) (*RoutineTaskPurpose, error) {
	for _, enumValue := range AllRoutineTaskPurposes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineTaskPurpose: %s", enumString)
}
