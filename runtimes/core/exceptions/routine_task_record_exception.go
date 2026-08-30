package apiexceptions

type RoutineTaskRecordException struct {
	Exception
}

func NewRoutineTaskRecordException() RoutineTaskRecordException {
	return RoutineTaskRecordException{Exception: NewException("RoutineTaskRecord")}
}
