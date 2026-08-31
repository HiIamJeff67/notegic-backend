package routinetask

import "testing"

func TestNewManagerCreatesPreparationService(t *testing.T) {
	manager := NewManager(1)

	if manager.preparationService == nil {
		t.Fatal("routine task preparation service is nil")
	}
}
