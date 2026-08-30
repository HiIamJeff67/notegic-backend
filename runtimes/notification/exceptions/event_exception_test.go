package exceptions

import "testing"

func TestEventUnsupportedType(t *testing.T) {
	exception := NewEventException("Notification").UnsupportedEventType()
	if exception.Reason != "UnsupportedEventType" || exception.Domain != "Notification" {
		t.Fatalf("unexpected event exception: %#v", exception)
	}
}
