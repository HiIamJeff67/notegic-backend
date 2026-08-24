package enums

type RoutineStatus string

const (
	RoutineStatus_Scheduled  RoutineStatus = "Scheduled"
	RoutineStatus_InProgress RoutineStatus = "InProgress"
	RoutineStatus_Completed  RoutineStatus = "Completed"
	RoutineStatus_OverDue    RoutineStatus = "OverDue"
)
