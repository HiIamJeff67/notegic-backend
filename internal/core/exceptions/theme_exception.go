package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type ThemeException struct {
	Exception
}

func NewThemeException() ThemeException {
	return ThemeException{Exception: NewException("Theme")}
}

func (ThemeException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"Theme",
		"Repository",
		"Theme was not found",
		http.StatusNotFound,
	)
}

func (ThemeException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"Theme",
		"Repository",
		"Failed to create Theme",
		http.StatusInternalServerError,
		true,
	)
}

func (ThemeException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"Theme",
		"Repository",
		"Failed to update Theme",
		http.StatusInternalServerError,
		true,
	)
}

func (ThemeException) FailedToDelete() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToDelete",
		"Theme",
		"Repository",
		"Failed to delete Theme",
		http.StatusInternalServerError,
		true,
	)
}

func (ThemeException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"Theme",
		"Repository",
		"No changes were applied to Theme",
		http.StatusNotModified,
	)
}
