package dependencies

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidateAcceptsAcyclicDependencies(t *testing.T) {
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	thirdTaskId := uuid.New()

	err := Validate(
		[]uuid.UUID{firstTaskId, secondTaskId, thirdTaskId},
		[]Edge{
			{TaskId: secondTaskId, PreviousTaskId: firstTaskId},
			{TaskId: thirdTaskId, PreviousTaskId: secondTaskId},
		},
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsDependencyCycles(t *testing.T) {
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()

	err := Validate(
		[]uuid.UUID{firstTaskId, secondTaskId},
		[]Edge{
			{TaskId: firstTaskId, PreviousTaskId: secondTaskId},
			{TaskId: secondTaskId, PreviousTaskId: firstTaskId},
		},
	)
	if err == nil || !strings.Contains(err.Error(), "contains a cycle") {
		t.Fatalf("Validate() error = %v, want a cycle error", err)
	}
}
