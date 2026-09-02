package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RoutineTaskDependencyException struct {
	CoreException
}

func NewRoutineTaskDependencyException() RoutineTaskDependencyException {
	return RoutineTaskDependencyException{
		CoreException: CoreException{
			Domain: "RoutineTaskDependency",
		},
	}
}

func (RoutineTaskDependencyException) DependencyAlreadyExists() *cexceptions.Exception {
	return cexceptions.New(
		"DependencyAlreadyExists",
		"RoutineTaskDependency",
		"Create",
		"The routine task dependency already exists",
		http.StatusConflict,
	)
}
