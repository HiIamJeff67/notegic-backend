package exceptions

type RoutineTaskRecordException struct {
	CoreException
}

func NewRoutineTaskRecordException() RoutineTaskRecordException {
	return RoutineTaskRecordException{
		CoreException: CoreException{
			Domain: "RoutineTaskRecord",
		},
	}
}
