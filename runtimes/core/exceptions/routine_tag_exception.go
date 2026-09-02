package exceptions

type RoutineTagException struct {
	CoreException
}

func NewRoutineTagException() RoutineTagException {
	return RoutineTagException{
		CoreException: CoreException{
			Domain: "RoutineTag",
		},
	}
}
