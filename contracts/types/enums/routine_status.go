package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineStatus string

const (
	RoutineStatus_Scheduled  RoutineStatus = "Scheduled"
	RoutineStatus_InProgress RoutineStatus = "InProgress"
	RoutineStatus_Completed  RoutineStatus = "Completed"
	RoutineStatus_OverDue    RoutineStatus = "OverDue"
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

func (value RoutineStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineStatus(string(v))
		return nil
	case string:
		*value = RoutineStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineStatus) String() string {
	return string(value)
}

func (value *RoutineStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineStatuses, *value)
}

func ConvertStringToRoutineStatus(enumString string) (*RoutineStatus, error) {
	for _, enumValue := range AllRoutineStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineStatus: %s", enumString)
}
