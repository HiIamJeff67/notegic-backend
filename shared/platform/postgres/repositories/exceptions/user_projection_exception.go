package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UserProjectionException struct {
	RepositoryException
}

func NewUserProjectionException() UserProjectionException {
	return UserProjectionException{
		RepositoryException: NewRepositoryException("UserProjection"),
	}
}

func (e UserProjectionException) FailedToCreate() *cexceptions.Exception {
	return e.New(
		"FailedToCreate",
		"Create",
		"Failed to create user projection",
		http.StatusInternalServerError,
		true,
	)
}
