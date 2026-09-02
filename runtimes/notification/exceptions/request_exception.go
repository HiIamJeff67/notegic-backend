package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RequestException struct {
	NotificationException
}

func NewRequestException(domain string) RequestException {
	return RequestException{
		NotificationException: NotificationException{
			Domain: domain,
		},
	}
}

func (e RequestException) RecipientRequired() *cexceptions.Exception {
	return cexceptions.New("RecipientUserPublicIdRequired", e.Domain, "ValidateRequest", "The recipient user public ID is required", http.StatusBadRequest)
}

func (e RequestException) InvalidSearchRequest(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidSearchRequest", e.Domain, "SearchPrivateNotifications", "The private notification search request is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e RequestException) InvalidCountRequest(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidCountRequest", e.Domain, "CountUnreadNotifications", "The unread notification count request is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e RequestException) InvalidMarkReadRequest(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidMarkReadRequest", e.Domain, "MarkNotificationsRead", "The mark-read request is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e RequestException) InvalidDeleteRequest(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidDeleteRequest", e.Domain, "DeleteNotifications", "The delete notifications request is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e RequestException) UserRequired() *cexceptions.Exception {
	return cexceptions.New("UserPublicIdRequired", e.Domain, "DeleteAllNotificationsForUser", "The user public ID is required", http.StatusBadRequest)
}
