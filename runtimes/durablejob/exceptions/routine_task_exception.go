package exceptions

import (
	"fmt"
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RoutineTaskException struct {
	DurableJobException
}

func NewRoutineTaskException() RoutineTaskException {
	return RoutineTaskException{
		DurableJobException: DurableJobException{
			Domain: "RoutineTask",
		},
	}
}

func (e RoutineTaskException) InvalidPayload(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"InvalidRoutineTaskPayload",
		e.Domain,
		"PrepareRoutineTask",
		"The routine task payload is invalid",
		http.StatusBadRequest,
	).WithOrigin(cause)
}

func (e RoutineTaskException) Canceled(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"Canceled",
		e.Domain,
		"ExecuteRoutineTask",
		"The routine task was canceled",
		http.StatusRequestTimeout,
	).WithOrigin(cause)
}

func (e RoutineTaskException) Timeout(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"Timeout",
		e.Domain,
		"ExecuteRoutineTask",
		"The routine task timed out",
		http.StatusRequestTimeout,
		true,
	).WithOrigin(cause)
}

func (e RoutineTaskException) TargetNotFound(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"TargetNotFound",
		e.Domain,
		"ExecuteRoutineTask",
		"The routine task target was not found",
		http.StatusNotFound,
	).WithOrigin(cause)
}

func (e RoutineTaskException) PermissionDenied(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"PermissionDenied",
		e.Domain,
		"ExecuteRoutineTask",
		"The routine task is not permitted",
		http.StatusForbidden,
	).WithOrigin(cause)
}

func (e RoutineTaskException) HandlerFailed(cause error) *cexceptions.Exception {
	return cexceptions.New(
		"HandlerFailed",
		e.Domain,
		"ExecuteRoutineTask",
		fmt.Sprintf("The routine task handler failed: %v", cause),
		http.StatusInternalServerError,
		true,
	).WithOrigin(cause)
}
