package enums

import "testing"

func TestRoutineTaskPurposesAreUniqueAndComplete(t *testing.T) {
	if got, want := len(AllRoutineTaskPurposes), 16; got != want {
		t.Fatalf("current purpose count = %d, want %d", got, want)
	}

	seen := make(map[RoutineTaskPurpose]struct{}, len(AllRoutineTaskPurposes))
	for _, purpose := range AllRoutineTaskPurposes {
		if _, exists := seen[purpose]; exists {
			t.Fatalf("duplicate current purpose %q", purpose)
		}
		seen[purpose] = struct{}{}
	}
}

func TestRoutineTaskPurposesHaveObjectKinds(t *testing.T) {
	for _, purpose := range AllRoutineTaskPurposes {
		kind, ok := purpose.ObjectKind()
		if !ok {
			t.Fatalf("purpose %q has no object kind", purpose)
		}
		if purpose == RoutineTaskPurpose_GetSubShelf || purpose == RoutineTaskPurpose_CreateSubShelf ||
			purpose == RoutineTaskPurpose_UpdateSubShelf || purpose == RoutineTaskPurpose_DeleteSubShelf ||
			purpose == RoutineTaskPurpose_GetRoutine || purpose == RoutineTaskPurpose_CreateRoutine ||
			purpose == RoutineTaskPurpose_UpdateRoutine || purpose == RoutineTaskPurpose_DeleteRoutine {
			if kind != RoutineTaskObjectKind_Container {
				t.Fatalf("purpose %q kind = %q, want %q", purpose, kind, RoutineTaskObjectKind_Container)
			}
			continue
		}
		if kind != RoutineTaskObjectKind_Core {
			t.Fatalf("purpose %q kind = %q, want %q", purpose, kind, RoutineTaskObjectKind_Core)
		}
	}
}
