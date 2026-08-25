package apiexceptions

type RoutineTaskException struct {
	Exception
}

func NewRoutineTaskException() RoutineTaskException {
	return RoutineTaskException{
		Exception: NewException("RoutineTask"),
	}
}
