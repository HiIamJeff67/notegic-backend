package exceptions

type NotificationException struct {
	Domain string
}

func NewNotificationException(domain string) NotificationException {
	return NotificationException{Domain: domain}
}
