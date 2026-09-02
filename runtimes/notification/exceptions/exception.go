package exceptions

type NotificationException struct {
	Domain string
}

func NewNotificationException() NotificationException {
	return NotificationException{
		Domain: "Notification",
	}
}
