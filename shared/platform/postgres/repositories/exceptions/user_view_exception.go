package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UserViewException struct {
	RepositoryException
}

func NewUserViewException() UserViewException {
	return UserViewException{RepositoryException: NewRepositoryException("UserView")}
}

func (e UserViewException) NotFoundByPublicId() *cexceptions.Exception {
	return e.New(
		"NotFound",
		"GetOneByPublicId",
		"User view not found",
		http.StatusNotFound,
	)
}

func (e UserViewException) FailedToGetByPublicId() *cexceptions.Exception {
	return e.New(
		"FailedToGet",
		"GetOneByPublicId",
		"Failed to get user view",
		http.StatusInternalServerError,
		true,
	)
}
