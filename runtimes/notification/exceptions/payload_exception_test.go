package exceptions

import "testing"

func TestPayloadDecodeFailed(t *testing.T) {
	cause := &testError{message: "invalid payload"}
	exception := NewPayloadException("Notification").PayloadDecodeFailed(cause)
	if exception.Reason != "PayloadDecodeFailed" || exception.Origin() != cause {
		t.Fatalf("unexpected payload exception: %#v", exception)
	}
}

type testError struct{ message string }

func (e *testError) Error() string { return e.message }
