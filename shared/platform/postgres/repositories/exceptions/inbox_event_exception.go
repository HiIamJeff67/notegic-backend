package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type InboxEventException struct {
	RepositoryException
}

func NewInboxEventException() InboxEventException {
	return InboxEventException{RepositoryException: NewRepositoryException("InboxEvent")}
}

func (e InboxEventException) EventIdRequired() *cexceptions.Exception {
	return e.New(
		"InvalidInput",
		"Create",
		"Inbox event ID is required",
		http.StatusBadRequest,
	)
}

func (e InboxEventException) FailedToRecord() *cexceptions.Exception {
	return e.New(
		"FailedToCreate",
		"Create",
		"Failed to record inbox event",
		http.StatusInternalServerError,
		true,
	)
}
