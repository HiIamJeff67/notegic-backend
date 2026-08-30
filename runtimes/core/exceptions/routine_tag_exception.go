package apiexceptions

type RoutineTagException struct {
	Exception
}

func NewRoutineTagException() RoutineTagException {
	return RoutineTagException{
		Exception: NewException("RoutineTag"),
	}
}
