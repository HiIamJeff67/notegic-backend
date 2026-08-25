package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type BlockException struct {
	CoreException
}

func NewBlockException() BlockException {
	return BlockException{CoreException: NewCoreException("Block")}
}

func (BlockException) NotFound() *cexceptions.Exception {
	return cexceptions.New(
		"NotFound",
		"Block",
		"Repository",
		"Block was not found",
		http.StatusNotFound,
	)
}

func (BlockException) FailedToCreate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCreate",
		"Block",
		"Repository",
		"Failed to create the block",
		http.StatusInternalServerError,
		true,
	)
}

func (BlockException) FailedToUpdate() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToUpdate",
		"Block",
		"Repository",
		"Failed to update the block",
		http.StatusInternalServerError,
		true,
	)
}

func (BlockException) NoChanges() *cexceptions.Exception {
	return cexceptions.New(
		"NoChanges",
		"Block",
		"Repository",
		"No changes were applied to the block",
		http.StatusNotModified,
	)
}

func (BlockException) FailedToCommitTransaction() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToCommitTransaction",
		"Block",
		"Transaction",
		"Failed to commit the block transaction",
		http.StatusInternalServerError,
		true,
	)
}

func (BlockException) InvalidDto() *cexceptions.Exception {
	return cexceptions.New(
		"InvalidDto",
		"Block",
		"Validate",
		"Invalid block DTO",
		http.StatusBadRequest,
	)
}

func (BlockException) NoPermission(action string) *cexceptions.Exception {
	return cexceptions.New(
		"PermissionDenied",
		"Block",
		"Authorize",
		"Permission is denied to "+action,
		http.StatusForbidden,
	)
}
