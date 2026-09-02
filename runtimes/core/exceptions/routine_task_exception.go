package exceptions

type RoutineTaskException struct {
	CoreException
}

func NewRoutineTaskException() RoutineTaskException {
	return RoutineTaskException{
		CoreException: CoreException{
			Domain: "RoutineTask",
		},
	}
}
