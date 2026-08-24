package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineTaskRecordStatus enumcontract.RoutineTaskRecordStatus

func (value *RoutineTaskRecordStatus) ToContractable() *enumcontract.RoutineTaskRecordStatus {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.RoutineTaskRecordStatus(*value)
	return &contractValue
}

func RoutineTaskRecordStatusToStorable(value *enumcontract.RoutineTaskRecordStatus) *RoutineTaskRecordStatus {
	if value == nil {
		return nil
	}

	storableValue := RoutineTaskRecordStatus(*value)
	return &storableValue
}

const (
	RoutineTaskRecordStatus_Running RoutineTaskRecordStatus = RoutineTaskRecordStatus(enumcontract.RoutineTaskRecordStatus_Running)
	RoutineTaskRecordStatus_Success RoutineTaskRecordStatus = RoutineTaskRecordStatus(enumcontract.RoutineTaskRecordStatus_Success)
	RoutineTaskRecordStatus_Failed  RoutineTaskRecordStatus = RoutineTaskRecordStatus(enumcontract.RoutineTaskRecordStatus_Failed)
	RoutineTaskRecordStatus_Cancel  RoutineTaskRecordStatus = RoutineTaskRecordStatus(enumcontract.RoutineTaskRecordStatus_Cancel)
)

var AllRoutineTaskRecordStatuses = []RoutineTaskRecordStatus{
	RoutineTaskRecordStatus_Running,
	RoutineTaskRecordStatus_Success,
	RoutineTaskRecordStatus_Failed,
	RoutineTaskRecordStatus_Cancel,
}

var AllRoutineTaskRecordStatusStrings = []string{
	string(RoutineTaskRecordStatus_Running),
	string(RoutineTaskRecordStatus_Success),
	string(RoutineTaskRecordStatus_Failed),
	string(RoutineTaskRecordStatus_Cancel),
}

func (rtrs RoutineTaskRecordStatus) Name() string {
	return reflect.TypeOf(rtrs).Name()
}

func (rtrs *RoutineTaskRecordStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*rtrs = RoutineTaskRecordStatus(string(v))
		return nil
	case string:
		*rtrs = RoutineTaskRecordStatus(v)
		return nil
	}
	return scanError(value, rtrs)
}

func (rtrs RoutineTaskRecordStatus) Value() (driver.Value, error) {
	return string(rtrs), nil
}

func (rtrs RoutineTaskRecordStatus) String() string {
	return string(rtrs)
}

func (rtrs *RoutineTaskRecordStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskRecordStatuses, *rtrs)
}

func ConvertStringToRoutineTaskRecordStatus(enumString string) (*RoutineTaskRecordStatus, error) {
	for _, routineTaskRecordStatus := range AllRoutineTaskRecordStatuses {
		if string(routineTaskRecordStatus) == enumString {
			return &routineTaskRecordStatus, nil
		}
	}
	return nil, fmt.Errorf("invalid routine task record status: %s", enumString)
}
