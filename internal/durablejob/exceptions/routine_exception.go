package exceptions

import (
	"fmt"
	"net/http"
	"time"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RoutineException struct {
	CoreException
}

func NewRoutineException() RoutineException {
	return RoutineException{
		CoreException: NewCoreException("Routine"),
	}
}

func (RoutineException) QueriedTimeRangeTooLarge(from time.Time, to time.Time) *cexceptions.Exception {
	return cexceptions.New(
		"QueriedTimeRangeTooLarge",
		"Routine",
		"Search",
		fmt.Sprintf("Cannot query the time range from %s to %s because it is too large", from, to),
		http.StatusBadRequest,
	)
}

func (RoutineException) FailedToLinkRoutineTags() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToLinkRoutineTags",
		"Routine",
		"Link",
		"Cannot link the given routine tags to the target routine",
		http.StatusBadRequest,
	)
}

func (RoutineException) FailedToLinkItems() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToLinkItems",
		"Routine",
		"Link",
		"Cannot link the given items to the target routine",
		http.StatusBadRequest,
	)
}
