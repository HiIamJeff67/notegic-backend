package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineTaskRecordErrorCode enumcontract.RoutineTaskRecordErrorCode

func (value *RoutineTaskRecordErrorCode) ToContractable() *enumcontract.RoutineTaskRecordErrorCode {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.RoutineTaskRecordErrorCode(*value)
	return &contractValue
}

func RoutineTaskRecordErrorCodeToStorable(value *enumcontract.RoutineTaskRecordErrorCode) *RoutineTaskRecordErrorCode {
	if value == nil {
		return nil
	}

	storableValue := RoutineTaskRecordErrorCode(*value)
	return &storableValue
}

const (
	RoutineTaskRecordErrorCode_PermissionDenied  RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_PermissionDenied)
	RoutineTaskRecordErrorCode_PayloadInvalid    RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_PayloadInvalid)
	RoutineTaskRecordErrorCode_TargetNotFound    RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_TargetNotFound)
	RoutineTaskRecordErrorCode_PlanLimitExceeded RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_PlanLimitExceeded)
	RoutineTaskRecordErrorCode_HandlerFailed     RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_HandlerFailed)
	RoutineTaskRecordErrorCode_DatabaseError     RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_DatabaseError)
	RoutineTaskRecordErrorCode_Timeout           RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_Timeout)
	RoutineTaskRecordErrorCode_Canceled          RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_Canceled)
	RoutineTaskRecordErrorCode_Unknown           RoutineTaskRecordErrorCode = RoutineTaskRecordErrorCode(enumcontract.RoutineTaskRecordErrorCode_Unknown)
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

func (rtrec RoutineTaskRecordErrorCode) Name() string {
	return reflect.TypeOf(rtrec).Name()
}

func (rtrec *RoutineTaskRecordErrorCode) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*rtrec = RoutineTaskRecordErrorCode(string(v))
		return nil
	case string:
		*rtrec = RoutineTaskRecordErrorCode(v)
		return nil
	}
	return scanError(value, rtrec)
}

func (rtrec RoutineTaskRecordErrorCode) Value() (driver.Value, error) {
	return string(rtrec), nil
}

func (rtrec RoutineTaskRecordErrorCode) String() string {
	return string(rtrec)
}

func (rtrec *RoutineTaskRecordErrorCode) IsValidEnum() bool {
	return slices.Contains(AllRoutineTaskRecordErrorCodes, *rtrec)
}

func ConvertStringToRoutineTaskRecordErrorCode(enumString string) (*RoutineTaskRecordErrorCode, error) {
	for _, routineTaskRecordErrorCode := range AllRoutineTaskRecordErrorCodes {
		if string(routineTaskRecordErrorCode) == enumString {
			return &routineTaskRecordErrorCode, nil
		}
	}
	return nil, fmt.Errorf("invalid routine task record error code: %s", enumString)
}
