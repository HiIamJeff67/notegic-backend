package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type OutboxException struct {
	RepositoryException
}

func NewOutboxException() OutboxException {
	return OutboxException{RepositoryException: NewRepositoryException("Outbox")}
}

func (e OutboxException) FailedToClaim() *cexceptions.Exception {
	return e.New(
		"FailedToGet",
		"Claim",
		"Failed to claim available outbox events",
		http.StatusInternalServerError,
		true,
	)
}

func (e OutboxException) FailedToCommitClaimTransaction() *cexceptions.Exception {
	return e.New(
		"TransactionCommitFailed",
		"Claim",
		"Failed to commit the outbox claim transaction",
		http.StatusInternalServerError,
		true,
	)
}

func (e OutboxException) FailedToMarkPublished() *cexceptions.Exception {
	return e.New(
		"FailedToUpdate",
		"MarkPublished",
		"Failed to mark outbox events as published",
		http.StatusInternalServerError,
		true,
	)
}

func (e OutboxException) FailedToMarkFailed() *cexceptions.Exception {
	return e.New(
		"FailedToUpdate",
		"MarkFailed",
		"Failed to schedule outbox event retries",
		http.StatusInternalServerError,
		true,
	)
}

func (e OutboxException) FailedToCleanup() *cexceptions.Exception {
	return e.New(
		"FailedToDelete",
		"Cleanup",
		"Failed to delete published outbox events",
		http.StatusInternalServerError,
		true,
	)
}
