package apiexceptions

import (
	"fmt"
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type ShelfException struct {
	Exception
}

func NewShelfException() ShelfException {
	return ShelfException{
		Exception: NewException("Shelf"),
	}
}

func (ShelfException) DuplicateName(name string) *cexceptions.Exception {
	return cexceptions.New(
		"DuplicateName",
		"Shelf",
		"Create",
		fmt.Sprintf("The name of %s is already in use", name),
		http.StatusConflict,
	)
}

func (ShelfException) MaximumDepthExceeded(currentDepth int32, maxDepth int32) *cexceptions.Exception {
	return cexceptions.New(
		"MaximumDepthExceeded",
		"Shelf",
		"Validate",
		fmt.Sprintf("The current depth of %d exceeds the maximum depth of %d", currentDepth, maxDepth),
		http.StatusBadRequest,
	)
}

func (ShelfException) InsertParentIntoItsChildren(destination any, target any) *cexceptions.Exception {
	return cexceptions.New(
		"InsertParentIntoItsChildren",
		"Shelf",
		"Validate",
		fmt.Sprintf("Cannot insert parent %v into child %v", target, destination),
		http.StatusBadRequest,
	)
}
