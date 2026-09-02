package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type EventException struct {
	NotificationException
}

func NewEventException(domain string) EventException {
	return EventException{
		NotificationException: NotificationException{
			Domain: domain,
		},
	}
}

func (e EventException) UnsupportedEventType() *cexceptions.Exception {
	return cexceptions.New("UnsupportedEventType", e.Domain, "ConsumeEvent", "The notification event type is unsupported", http.StatusBadRequest)
}

func (e EventException) AggregateRecipientMismatch() *cexceptions.Exception {
	return cexceptions.New("AggregateRecipientMismatch", e.Domain, "ConsumeEvent", "The notification aggregate recipient does not match", http.StatusBadRequest)
}

func (e EventException) InvalidMetadata(cause error) *cexceptions.Exception {
	return cexceptions.New("InvalidNotificationMetadata", e.Domain, "ValidateEvent", "The notification metadata is invalid", http.StatusBadRequest).WithOrigin(cause)
}

func (e EventException) UnsupportedTemplateVersion(cause error) *cexceptions.Exception {
	return cexceptions.New("UnsupportedTemplateVersion", e.Domain, "ValidateEvent", "The notification template version is unsupported", http.StatusBadRequest).WithOrigin(cause)
}

func (e EventException) InvalidNewsTemplateKey() *cexceptions.Exception {
	return cexceptions.New("InvalidNewsTemplateKey", e.Domain, "ValidateEvent", "The news template key is invalid", http.StatusBadRequest)
}

func (e EventException) InvalidWarningTemplateKey() *cexceptions.Exception {
	return cexceptions.New("InvalidWarningTemplateKey", e.Domain, "ValidateEvent", "The warning template key is invalid", http.StatusBadRequest)
}

func (e EventException) InvalidImportantTemplateKey() *cexceptions.Exception {
	return cexceptions.New("InvalidImportantTemplateKey", e.Domain, "ValidateEvent", "The important template key is invalid", http.StatusBadRequest)
}

func (e EventException) UnsupportedType(cause error) *cexceptions.Exception {
	return cexceptions.New("UnsupportedNotificationType", e.Domain, "ValidateEvent", "The notification type is unsupported", http.StatusBadRequest).WithOrigin(cause)
}
