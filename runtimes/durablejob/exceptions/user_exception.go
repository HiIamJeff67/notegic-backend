package exceptions

import (
	"fmt"
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type UserException struct {
	Exception
}

func NewUserException() UserException {
	return UserException{Exception: NewException("User")}
}

func (UserException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"User",
		"Repository",
		"User was not found",
		http.StatusNotFound,
	)
}

func (UserException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"User",
		"Repository",
		"Failed to create the user",
		http.StatusInternalServerError,
		true,
	)
}

func (UserException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"User",
		"Repository",
		"Failed to update the user",
		http.StatusInternalServerError,
		true,
	)
}

func (UserException) FailedToDelete() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToDelete",
		"User",
		"Repository",
		"Failed to delete the user",
		http.StatusInternalServerError,
		true,
	)
}

func (UserException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"User",
		"Repository",
		"No changes were applied to the user",
		http.StatusNotModified,
	)
}

func (UserException) FailedToCommitTransaction() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCommitTransaction",
		"User",
		"Transaction",
		"Failed to commit the user transaction",
		http.StatusInternalServerError,
		true,
	)
}

func (UserException) InvalidInput() *cexceptions.Exception {
	return cexceptions.New(
		"InvalidInput",
		"User",
		"Validate",
		"Invalid user input",
		http.StatusBadRequest,
	)
}

func (UserException) DuplicateName(name string) *cexceptions.Exception {
	return cexceptions.New(
		"DuplicateName",
		"User",
		"Create",
		fmt.Sprintf("The name of %s is already in use", name),
		http.StatusConflict,
	)
}

func (UserException) DuplicateEmail(email string) *cexceptions.Exception {
	return cexceptions.New(
		"DuplicateEmail",
		"User",
		"Create",
		fmt.Sprintf("The email of %s is already in use", email),
		http.StatusConflict,
	)
}
