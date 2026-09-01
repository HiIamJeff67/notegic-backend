package routinetask

import "testing"

func TestNewManagerCreatesAssignmentPreparer(t *testing.T) {
	manager := NewManager(1)

	if manager.preparer == nil {
		t.Fatal("routine task assignment preparer is nil")
	}
}
