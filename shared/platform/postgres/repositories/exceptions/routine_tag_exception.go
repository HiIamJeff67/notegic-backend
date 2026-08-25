package exceptions

type RoutineTagException struct {
	RepositoryException
}

func NewRoutineTagException() RoutineTagException {
	return RoutineTagException{RepositoryException: NewRepositoryException("RoutineTag")}
}
