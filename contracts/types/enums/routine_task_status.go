package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineTaskStatus string

const (
	RoutineTaskStatus_Idle    RoutineTaskStatus = "Idle"
	RoutineTaskStatus_Waiting RoutineTaskStatus = "Waiting" // include scheduling, but we don't need to present to the client
	RoutineTaskStatus_Running RoutineTaskStatus = "Running"
	RoutineTaskStatus_Pause   RoutineTaskStatus = "Pause"
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

func (value RoutineTaskStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineTaskStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineTaskStatus(string(v))
		return nil
	case string:
		*value = RoutineTaskStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineTaskStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineTaskStatus) String() string {
	return string(value)
}

func (value *RoutineTaskStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskStatuses, *value)
}

func ConvertStringToRoutineTaskStatus(enumString string) (*RoutineTaskStatus, error) {
	for _, enumValue := range AllRoutineTaskStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineTaskStatus: %s", enumString)
}
