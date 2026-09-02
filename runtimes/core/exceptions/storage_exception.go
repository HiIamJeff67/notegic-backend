package exceptions

import (
	"fmt"
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type StorageException struct {
	CoreException
}

func NewStorageException() StorageException {
	return StorageException{
		CoreException: CoreException{
			Domain: "Storage",
		},
	}
}

func (StorageException) FailedToReadObjectBytes() *cexceptions.Exception {
	return cexceptions.New(
		"FailedToReadObjectBytes",
		"Storage",
		"Read",
		"Failed to read object bytes",
		http.StatusInternalServerError,
		true,
	)
}

func (StorageException) FailedToPutObject(object any) *cexceptions.Exception {
	return cexceptions.New(
		"FailedToPutObject",
		"Storage",
		"Put",
		fmt.Sprintf("Failed to put object %v", object),
		http.StatusInternalServerError,
		true,
	)
}

func (StorageException) FailedToPresignedGetObject(object any) *cexceptions.Exception {
	return cexceptions.New(
		"FailedToPresignedGetObject",
		"Storage",
		"PresignGet",
		fmt.Sprintf("Failed to presign object %v", object),
		http.StatusInternalServerError,
		true,
	)
}
