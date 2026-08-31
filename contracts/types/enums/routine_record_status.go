package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineRecordStatus string

const (
	RoutineRecordStatus_Pending  RoutineRecordStatus = "Pending"
	RoutineRecordStatus_Running  RoutineRecordStatus = "Running"
	RoutineRecordStatus_Success  RoutineRecordStatus = "Success"
	RoutineRecordStatus_Failed   RoutineRecordStatus = "Failed"
	RoutineRecordStatus_Blocked  RoutineRecordStatus = "Blocked"
	RoutineRecordStatus_Canceled RoutineRecordStatus = "Canceled"
)

var AllRoutineRecordStatuses = []RoutineRecordStatus{
	RoutineRecordStatus_Pending,
	RoutineRecordStatus_Running,
	RoutineRecordStatus_Success,
	RoutineRecordStatus_Failed,
	RoutineRecordStatus_Blocked,
	RoutineRecordStatus_Canceled,
}

var AllRoutineRecordStatusStrings = []string{
	string(RoutineRecordStatus_Pending),
	string(RoutineRecordStatus_Running),
	string(RoutineRecordStatus_Success),
	string(RoutineRecordStatus_Failed),
	string(RoutineRecordStatus_Blocked),
	string(RoutineRecordStatus_Canceled),
}

func (value RoutineRecordStatus) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineRecordStatus) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineRecordStatus(string(v))
		return nil
	case string:
		*value = RoutineRecordStatus(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineRecordStatus) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineRecordStatus) String() string {
	return string(value)
}

func (value *RoutineRecordStatus) IsValidEnum() bool {
	return slices.Contains(AllRoutineRecordStatuses, *value)
}

func ConvertStringToRoutineRecordStatus(enumString string) (*RoutineRecordStatus, error) {
	for _, enumValue := range AllRoutineRecordStatuses {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineRecordStatus: %s", enumString)
}
