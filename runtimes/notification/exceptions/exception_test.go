package exceptions

import "testing"

func TestNewNotificationException(t *testing.T) {
	exception := NewNotificationException("Notification")
	if exception.Domain != "Notification" {
		t.Fatalf("domain = %q, want Notification", exception.Domain)
	}
}
