package exceptions

import "testing"

func TestDeliveryFailed(t *testing.T) {
	cause := &testError{message: "SMTP unavailable"}
	exception := NewDeliveryException("Email").DeliveryFailed(cause)
	if exception.Reason != "DeliveryFailed" || !exception.Retryable {
		t.Fatalf("unexpected delivery exception: %#v", exception)
	}
	if exception.Origin() != cause {
		t.Fatal("delivery exception does not preserve its origin")
	}
}

type testError struct{ message string }

func (e *testError) Error() string { return e.message }
