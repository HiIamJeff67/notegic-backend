package routinetasktypes

import (
	"testing"

	"github.com/google/uuid"
)

func TestRoutineTaskObjectReferenceDistinguishesFakeAndRealIds(t *testing.T) {
	fakeId := NewRoutineTaskFakeId()
	realId := RoutineTaskObjectReference(uuid.NewString())

	if !fakeId.IsFakeId() || fakeId.IsRealId() {
		t.Fatalf("fake id classification is invalid: %q", fakeId)
	}
	if !realId.IsRealId() || realId.IsFakeId() {
		t.Fatalf("real id classification is invalid: %q", realId)
	}
	if len(fakeId) >= len(realId) {
		t.Fatalf("fake id = %q, real id = %q; fake id should be shorter", fakeId, realId)
	}
}

func TestRoutineTaskObjectReferenceResolvesFacts(t *testing.T) {
	fakeId := RoutineTaskObjectReference("f_88888888888888888888888888888888")
	realId := uuid.New()

	resolvedId, err := fakeId.Resolve(map[string]uuid.UUID{string(fakeId): realId})
	if err != nil {
		t.Fatalf("resolve fake id: %v", err)
	}
	if resolvedId != realId {
		t.Fatalf("resolved id = %s, want %s", resolvedId, realId)
	}
	if _, err := fakeId.Resolve(nil); err == nil {
		t.Fatal("expected missing fact error")
	}
}
