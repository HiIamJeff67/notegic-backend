package routinetasktypes

import (
	"encoding/json"
	"testing"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestPreparedRoutineTaskContractRoundTripPreservesAttempt(t *testing.T) {
	prepared := PreparedRoutineTask{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Attempt:             2,
		Purpose:             cenums.RoutineTaskPurpose_GetMaterial,
		Payload:             json.RawMessage(`{"name":"daily"}`),
		PreparedAt:          time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
	}

	encoded, err := json.Marshal(prepared)
	if err != nil {
		t.Fatalf("marshal prepared task: %v", err)
	}

	var decoded PreparedRoutineTask
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal prepared task: %v", err)
	}
	if decoded.Attempt != prepared.Attempt || decoded.Purpose != prepared.Purpose {
		t.Fatalf("decoded contract = %#v", decoded)
	}
}

func TestCompletedRoutineTaskContractRequiresAttempt(t *testing.T) {
	request := CompletedRoutineTask{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		CompletedAt:         time.Now().UTC(),
		PreparedTask: &PreparedRoutineTask{
			RoutineTaskId:       uuid.New(),
			RoutineTaskRecordId: uuid.New(),
			RoutineId:           uuid.New(),
			ActorUserId:         uuid.New(),
			ActorUserPublicId:   uuid.New(),
			Purpose:             cenums.RoutineTaskPurpose_GetMaterial,
			Payload:             json.RawMessage(`{"name":"daily"}`),
			PreparedAt:          time.Now().UTC(),
		},
	}
	if err := validator.New().Struct(request); err == nil {
		t.Fatal("expected a prepared task without attempt to be rejected")
	}
}
