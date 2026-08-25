package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineTaskRecordStatus string

const (
	RoutineTaskRecordStatus_Running RoutineTaskRecordStatus = "Running"
	RoutineTaskRecordStatus_Success RoutineTaskRecordStatus = "Success"
	RoutineTaskRecordStatus_Failed  RoutineTaskRecordStatus = "Failed"
	RoutineTaskRecordStatus_Cancel  RoutineTaskRecordStatus = "Cancel"
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

func (value RoutineTaskRecordStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineTaskRecordStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineTaskRecordStatus(string(v))
		return nil
	case string:
		*value = RoutineTaskRecordStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineTaskRecordStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineTaskRecordStatus) String() string {
	return string(value)
}

func (value *RoutineTaskRecordStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskRecordStatuses, *value)
}

func ConvertStringToRoutineTaskRecordStatus(enumString string) (*RoutineTaskRecordStatus, error) {
	for _, enumValue := range AllRoutineTaskRecordStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineTaskRecordStatus: %s", enumString)
}
