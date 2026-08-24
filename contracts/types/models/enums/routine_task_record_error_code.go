package enums

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
