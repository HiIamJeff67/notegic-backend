package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UserAccountException struct {
	CoreException
}

func NewUserAccountException() UserAccountException {
	return UserAccountException{
		CoreException: CoreException{
			Domain: "UserAccount",
		},
	}
}

func (UserAccountException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"UserAccount",
		"Repository",
		"User account was not found",
		http.StatusNotFound,
	)
}

func (UserAccountException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"UserAccount",
		"Repository",
		"Failed to create the user account",
		http.StatusInternalServerError,
		true,
	)
}

func (UserAccountException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"UserAccount",
		"Repository",
		"Failed to update the user account",
		http.StatusInternalServerError,
		true,
	)
}

func (UserAccountException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"UserAccount",
		"Repository",
		"No changes were applied to the user account",
		http.StatusNotModified,
	)
}
