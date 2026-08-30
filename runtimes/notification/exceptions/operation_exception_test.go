package exceptions

import "testing"

func TestOperationCreateFailed(t *testing.T) {
	cause := &testError{message: "database unavailable"}
	exception := NewOperationException("Notification").CreateFailed(cause)
	if exception.Reason != "CreateFailed" || !exception.Retryable || exception.Origin() != cause {
		t.Fatalf("unexpected operation exception: %#v", exception)
	}
}
