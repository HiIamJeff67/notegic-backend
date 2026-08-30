package exceptions

import "testing"

func TestNewRoutineExceptionKeepsDomain(t *testing.T) {
	exception := NewRoutineTaskException()
	if exception.Domain != "RoutineTask" {
		t.Fatalf("domain = %q, want RoutineTask", exception.Domain)
	}
}
