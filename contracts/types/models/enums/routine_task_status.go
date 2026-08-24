package enums

type RoutineTaskStatus string

const (
	RoutineTaskStatus_Idle    RoutineTaskStatus = "Idle"
	RoutineTaskStatus_Waiting RoutineTaskStatus = "Waiting" // include scheduling, but we don't need to present to the client
	RoutineTaskStatus_Running RoutineTaskStatus = "Running"
	RoutineTaskStatus_Pause   RoutineTaskStatus = "Pause"
)
