package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type RoutineTaskRecordErrorCode string

const (
	RoutineTaskRecordErrorCode_PermissionDenied  RoutineTaskRecordErrorCode = "PermissionDenied"
	RoutineTaskRecordErrorCode_PayloadInvalid    RoutineTaskRecordErrorCode = "PayloadInvalid"
	RoutineTaskRecordErrorCode_TargetNotFound    RoutineTaskRecordErrorCode = "TargetNotFound"
	RoutineTaskRecordErrorCode_PlanLimitExceeded RoutineTaskRecordErrorCode = "PlanLimitExceeded"
	RoutineTaskRecordErrorCode_HandlerFailed     RoutineTaskRecordErrorCode = "HandlerFailed"
	RoutineTaskRecordErrorCode_DatabaseError     RoutineTaskRecordErrorCode = "DatabaseError"
	RoutineTaskRecordErrorCode_Timeout           RoutineTaskRecordErrorCode = "Timeout"
	RoutineTaskRecordErrorCode_Canceled          RoutineTaskRecordErrorCode = "Canceled"
	RoutineTaskRecordErrorCode_Unknown           RoutineTaskRecordErrorCode = "Unknown"
)

var AllRoutineTaskRecordErrorCodes = []RoutineTaskRecordErrorCode{
	RoutineTaskRecordErrorCode_PermissionDenied,
	RoutineTaskRecordErrorCode_PayloadInvalid,
	RoutineTaskRecordErrorCode_TargetNotFound,
	RoutineTaskRecordErrorCode_PlanLimitExceeded,
	RoutineTaskRecordErrorCode_HandlerFailed,
	RoutineTaskRecordErrorCode_DatabaseError,
	RoutineTaskRecordErrorCode_Timeout,
	RoutineTaskRecordErrorCode_Canceled,
	RoutineTaskRecordErrorCode_Unknown,
}

var AllRoutineTaskRecordErrorCodeStrings = []string{
	string(RoutineTaskRecordErrorCode_PermissionDenied),
	string(RoutineTaskRecordErrorCode_PayloadInvalid),
	string(RoutineTaskRecordErrorCode_TargetNotFound),
	string(RoutineTaskRecordErrorCode_PlanLimitExceeded),
	string(RoutineTaskRecordErrorCode_HandlerFailed),
	string(RoutineTaskRecordErrorCode_DatabaseError),
	string(RoutineTaskRecordErrorCode_Timeout),
	string(RoutineTaskRecordErrorCode_Canceled),
	string(RoutineTaskRecordErrorCode_Unknown),
}

func (value RoutineTaskRecordErrorCode) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *RoutineTaskRecordErrorCode) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = RoutineTaskRecordErrorCode(string(v))
		return nil
	case string:
		*value = RoutineTaskRecordErrorCode(v)
		return nil
	}
	return scanError(raw, value)
}

func (value RoutineTaskRecordErrorCode) Value() (driver.Value, error) {
	return string(value), nil
}

func (value RoutineTaskRecordErrorCode) String() string {
	return string(value)
}

func (value *RoutineTaskRecordErrorCode) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskRecordErrorCodes, *value)
}

func ConvertStringToRoutineTaskRecordErrorCode(enumString string) (*RoutineTaskRecordErrorCode, error) {
	for _, enumValue := range AllRoutineTaskRecordErrorCodes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid RoutineTaskRecordErrorCode: %s", enumString)
}
