package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineTaskStatus enumcontract.RoutineTaskStatus

func (value *RoutineTaskStatus) ToContractable() *enumcontract.RoutineTaskStatus {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.RoutineTaskStatus(*value)
	return &contractValue
}

func RoutineTaskStatusToStorable(value *enumcontract.RoutineTaskStatus) *RoutineTaskStatus {
	if value == nil {
		return nil
	}

	storableValue := RoutineTaskStatus(*value)
	return &storableValue
}

const (
	RoutineTaskStatus_Idle    RoutineTaskStatus = RoutineTaskStatus(enumcontract.RoutineTaskStatus_Idle)
	RoutineTaskStatus_Waiting RoutineTaskStatus = RoutineTaskStatus(enumcontract.RoutineTaskStatus_Waiting) // include scheduling, but we don't need to present to the client
	RoutineTaskStatus_Running RoutineTaskStatus = RoutineTaskStatus(enumcontract.RoutineTaskStatus_Running)
	RoutineTaskStatus_Pause   RoutineTaskStatus = RoutineTaskStatus(enumcontract.RoutineTaskStatus_Pause)
)

var AllRoutineTaskStatuses = []RoutineTaskStatus{
	RoutineTaskStatus_Idle,
	RoutineTaskStatus_Waiting,
	RoutineTaskStatus_Running,
	RoutineTaskStatus_Pause,
}

var AllRoutineTaskStatusStrings = []string{
	string(RoutineTaskStatus_Idle),
	string(RoutineTaskStatus_Waiting),
	string(RoutineTaskStatus_Running),
	string(RoutineTaskStatus_Pause),
}

func (rts RoutineTaskStatus) Name() string {
	return reflect.TypeOf(rts).Name()
}

func (rts *RoutineTaskStatus) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*rts = RoutineTaskStatus(string(v))
		return nil
	case string:
		*rts = RoutineTaskStatus(v)
		return nil
	}
	return scanError(value, rts)
}

func (rts RoutineTaskStatus) Value() (driver.Value, error) {
	return string(rts), nil
}

func (rts RoutineTaskStatus) String() string {
	return string(rts)
}

func (rts *RoutineTaskStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskStatuses, *rts)
}

func ConvertStringToRoutineTaskStatus(enumString string) (*RoutineTaskStatus, error) {
	for _, routineTaskStatus := range AllRoutineTaskStatuses {
		if string(routineTaskStatus) == enumString {
			return &routineTaskStatus, nil
		}
	}
	return nil, fmt.Errorf("invalid routine task status: %s", enumString)
}
