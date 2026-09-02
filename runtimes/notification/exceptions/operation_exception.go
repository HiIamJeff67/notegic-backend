package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type OperationException struct {
	NotificationException
}

func NewOperationException(domain string) OperationException {
	return OperationException{
		NotificationException: NotificationException{
			Domain: domain,
		},
	}
}

func (e OperationException) CreateFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("CreateFailed", e.Domain, "CreateNotification", "Failed to create the notification", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) SearchFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("SearchFailed", e.Domain, "SearchPrivateNotifications", "Failed to search private notifications", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) CountUnreadFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("CountUnreadFailed", e.Domain, "CountUnreadNotifications", "Failed to count notifications", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) MarkReadFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("MarkReadFailed", e.Domain, "MarkNotificationsRead", "Failed to mark notifications as read", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) DeleteFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("DeleteFailed", e.Domain, "DeleteNotifications", "Failed to delete notifications", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) HardDeleteFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("HardDeleteFailed", e.Domain, "HardDeleteNotifications", "Failed to hard-delete notifications", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e OperationException) DeleteAllForUserFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("DeleteAllForUserFailed", e.Domain, "DeleteAllNotificationsForUser", "Failed to delete notifications for the user", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}
