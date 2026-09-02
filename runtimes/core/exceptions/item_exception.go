package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type ItemException struct {
	CoreException
}

func NewItemException() ItemException {
	return ItemException{
		CoreException: CoreException{
			Domain: "Item",
		},
	}
}

func (ItemException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"Item",
		"Repository",
		"Item was not found",
		http.StatusNotFound,
	)
}

func (ItemException) NoPermission(_ ...string) *cexceptions.Exception {
	return cexceptions.New(
		"PermissionDenied",
		"Item",
		"Authorize",
		"Permission is denied",
		http.StatusBadRequest,
	)
}
