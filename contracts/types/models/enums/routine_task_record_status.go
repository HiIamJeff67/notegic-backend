package enums

type RoutineTaskRecordStatus string

const (
	RoutineTaskRecordStatus_Running RoutineTaskRecordStatus = "Running"
	RoutineTaskRecordStatus_Success RoutineTaskRecordStatus = "Success"
	RoutineTaskRecordStatus_Failed  RoutineTaskRecordStatus = "Failed"
	RoutineTaskRecordStatus_Cancel  RoutineTaskRecordStatus = "Cancel"
)
