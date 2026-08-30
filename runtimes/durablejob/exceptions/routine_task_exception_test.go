package exceptions

import (
	"errors"
	"testing"
)

func TestRoutineTaskInvalidPayload(t *testing.T) {
	cause := errors.New("invalid payload")
	exception := NewRoutineTaskException().InvalidPayload(cause)

	if exception.Reason != "InvalidRoutineTaskPayload" || exception.Domain != "RoutineTask" {
		t.Fatalf("unexpected routine task exception: %#v", exception)
	}
	if exception.Origin() != cause {
		t.Fatal("routine task exception does not preserve its cause")
	}
}
