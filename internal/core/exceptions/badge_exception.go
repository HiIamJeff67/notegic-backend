package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type BadgeException struct {
	CoreException
}

func NewBadgeException() BadgeException {
	return BadgeException{CoreException: NewCoreException("Badge")}
}

func (BadgeException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"Badge",
		"Repository",
		"Badge was not found",
		http.StatusNotFound,
	)
}
