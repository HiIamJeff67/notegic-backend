package exceptions

type RoutineException struct {
	RepositoryException
}

func NewRoutineException() RoutineException {
	return RoutineException{RepositoryException: NewRepositoryException("Routine")}
}
