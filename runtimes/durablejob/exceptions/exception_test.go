package exceptions

import "testing"

func TestNewDurableJobExceptionKeepsRuntimeDomain(t *testing.T) {
	exception := NewDurableJobException()
	if exception.Domain != "DurableJob" {
		t.Fatalf("domain = %q, want DurableJob", exception.Domain)
	}
}

func TestNewRoutineExceptionKeepsDomain(t *testing.T) {
	exception := NewRoutineTaskException()
	if exception.Domain != "RoutineTask" {
		t.Fatalf("domain = %q, want RoutineTask", exception.Domain)
	}
}
