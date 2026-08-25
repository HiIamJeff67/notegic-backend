package exceptions

type RoutineTaskException struct {
	RepositoryException
}

func NewRoutineTaskException() RoutineTaskException {
	return RoutineTaskException{RepositoryException: NewRepositoryException("RoutineTask")}
}
