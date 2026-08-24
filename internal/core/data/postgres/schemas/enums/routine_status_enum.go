package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineStatus enumcontract.RoutineStatus

func (value *RoutineStatus) ToContractable() *enumcontract.RoutineStatus {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.RoutineStatus(*value)
	return &contractValue
}

func RoutineStatusToStorable(value *enumcontract.RoutineStatus) *RoutineStatus {
	if value == nil {
		return nil
	}

	storableValue := RoutineStatus(*value)
	return &storableValue
}

const (
	RoutineStatus_Scheduled  RoutineStatus = RoutineStatus(enumcontract.RoutineStatus_Scheduled)
	RoutineStatus_InProgress RoutineStatus = RoutineStatus(enumcontract.RoutineStatus_InProgress)
	RoutineStatus_Completed  RoutineStatus = RoutineStatus(enumcontract.RoutineStatus_Completed)
	RoutineStatus_OverDue    RoutineStatus = RoutineStatus(enumcontract.RoutineStatus_OverDue)
)

var AllRoutineStatuses = []RoutineStatus{
	RoutineStatus_Scheduled,
	RoutineStatus_InProgress,
	RoutineStatus_Completed,
	RoutineStatus_OverDue,
}

var AllRoutineStatusStrings = []string{
	string(RoutineStatus_Scheduled),
	string(RoutineStatus_InProgress),
	string(RoutineStatus_Completed),
	string(RoutineStatus_OverDue),
}

func (rs RoutineStatus) Name() string {
	return reflect.TypeOf(rs).Name()
}

func (rs *RoutineStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*rs = RoutineStatus(string(v))
		return nil
	case string:
		*rs = RoutineStatus(v)
		return nil
	}
	return scanError(value, rs)
}

func (rs RoutineStatus) Value() (driver.Value, error) {
	return string(rs), nil
}

func (rs RoutineStatus) String() string {
	return string(rs)
}

func (rs *RoutineStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineStatuses, *rs)
}

func ConvertStringToRoutineStatus(enumString string) (*RoutineStatus, error) {
	for _, routineStatus := range AllRoutineStatuses {
		if string(routineStatus) == enumString {
			return &routineStatus, nil
		}
	}
	return nil, fmt.Errorf("invalid routine status: %s", enumString)
}
