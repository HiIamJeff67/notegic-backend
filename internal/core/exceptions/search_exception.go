package apiexceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type SearchException struct {
	Exception
}

func NewSearchException() SearchException {
	return SearchException{Exception: NewException("Search")}
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
