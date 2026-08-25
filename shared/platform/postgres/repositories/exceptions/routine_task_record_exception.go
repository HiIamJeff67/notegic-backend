package exceptions

type RoutineTaskRecordException struct {
	RepositoryException
}

func NewRoutineTaskRecordException() RoutineTaskRecordException {
	return RoutineTaskRecordException{RepositoryException: NewRepositoryException("RoutineTaskRecord")}
}
