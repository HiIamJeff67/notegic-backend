package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type PayloadException struct {
	NotificationException
}

func NewPayloadException(domain string) PayloadException {
	return PayloadException{
		NotificationException: NotificationException{
			Domain: domain,
		},
	}
}

func (e PayloadException) PayloadDecodeFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("PayloadDecodeFailed", e.Domain, "DecodePayload", "Failed to decode the notification payload", http.StatusBadRequest).WithOrigin(cause)
}

func (e PayloadException) InvalidNewsPayload(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidNewsPayload", e.Domain, "ValidatePayload", "The news notification payload is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e PayloadException) InvalidWarningPayload(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidWarningPayload", e.Domain, "ValidatePayload", "The warning notification payload is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e PayloadException) InvalidImportantPayload(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidImportantPayload", e.Domain, "ValidatePayload", "The important notification payload is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e PayloadException) ResponsePayloadDecodeFailed(cause error) *cexceptions.Exception {
	exception := cexceptions.New("ResponsePayloadDecodeFailed", e.Domain, "SearchPrivateNotifications", "Failed to decode a notification response payload", http.StatusInternalServerError, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}
