package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type SearchException struct {
	CoreException
}

func NewSearchException() SearchException {
	return SearchException{
		CoreException: CoreException{
			Domain: "Search",
		},
	}
}

func (SearchException) FailedToDecode() *cexceptions.Exception {
	return cexceptions.New(
		"CursorDecodeFailed",
		"Search",
		"Cursor",
		"Failed to decode the search cursor",
		http.StatusInternalServerError,
		true,
	)
}

func (SearchException) FailedToEncode() *cexceptions.Exception {
	return cexceptions.New(
		"CursorEncodeFailed",
		"Search",
		"Cursor",
		"Failed to encode the search cursor",
		http.StatusInternalServerError,
		true,
	)
}

func (SearchException) FailedToUnmarshalSearchCursor() *cexceptions.Exception {
	return cexceptions.New(
		"CursorEncodingFailed",
		"Search",
		"Cursor",
		"Failed to encode the search cursor",
		http.StatusInternalServerError,
		true,
	)
}
