package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	validation "github.com/HiIamJeff67/notegic-backend/internal/durablejob/validations"
)

func TestPurposeHandlerPreparesAssignmentWithoutDatabaseAccess(t *testing.T) {
	payload, err := json.Marshal(croutinetasktypes.CreateRootShelfRoutineTaskPayload{
		Name: "Daily {{date}}",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	assignment := croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             enums.RoutineTaskPurpose_CreateRootShelf,
		Payload:             payload,
		Attempt:             1,
		ScheduledAt:         time.Now().UTC(),
		StartedAt:           time.Now().UTC(),
		PatternValues:       map[string]string{"date": "2026-08-05"},
	}

	prepared, exception := NewPurposeHandler(validation.New()).HandlerFunc(t.Context(), assignment)
	if exception != nil {
		t.Fatalf("prepare assignment: %v", exception)
	}
	if prepared == nil || prepared.RoutineTaskId != assignment.RoutineTaskId {
		t.Fatalf("prepared task = %#v", prepared)
	}

	var preparedPayload croutinetasktypes.CreateRootShelfRoutineTaskPayload
	if err := json.Unmarshal(prepared.Payload, &preparedPayload); err != nil {
		t.Fatalf("decode prepared payload: %v", err)
	}
	if preparedPayload.Name != "Daily 2026-08-05" {
		t.Fatalf("prepared name = %q", preparedPayload.Name)
	}
}

func TestPurposeHandlerReturnsLocalErrorForInvalidPayload(t *testing.T) {
	assignment := croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             enums.RoutineTaskPurpose_CreateRootShelf,
		Payload:             []byte("{"),
	}

	prepared, err := NewPurposeHandler(validation.New()).HandlerFunc(t.Context(), assignment)
	if prepared != nil {
		t.Fatalf("prepared task = %#v, want nil", prepared)
	}
	if durableJobError, ok := err.(*cexceptions.Exception); !ok {
		t.Fatalf("error type = %T, want *exceptions.Exception", err)
	} else if durableJobError.Reason != "InvalidRoutineTaskPayload" || durableJobError.Domain != "RoutineTask" {
		t.Fatalf("error = %#v, want InvalidRoutineTaskPayload/RoutineTask", durableJobError)
	}
}

func TestPrepareAssignmentMatchesNestedTemplateBlockContent(t *testing.T) {
	payload, err := json.Marshal(croutinetasktypes.CreateBlockPackRoutineTaskPayload{
		TargetSubShelfId: uuid.New(),
		Template: croutinetasktypes.CreateBlockPackRoutineTaskTemplate{
			Name: "Daily note for {{date1}}",
			Blocks: []croutinetasktypes.CreateBlockPackRoutineTaskTemplateBlock{
				{
					ClientId: uuid.NewString(),
					ArborizedEditableBlock: cblocknote.ArborizedEditableBlock{
						Id:   uuid.New(),
						Type: enums.BlockType_Paragraph,
						Props: &cblocknote.BaseProps{
							Template: true,
						},
						Content: cblocknote.InlineContentList{
							{InlineContentUnion: cblocknote.NewStyledText("Daily note for {{date1}}", cblocknote.Styles{})},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	prepared, err := prepareAssignment(nil, nil, croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             enums.RoutineTaskPurpose_CreateBlockPack,
		Payload:             payload,
		PatternValues:       map[string]string{"date1": "2026-08-13"},
	})
	if err != nil {
		t.Fatalf("prepareAssignment() error = %v", err)
	}

	var preparedPayload croutinetasktypes.CreateBlockPackRoutineTaskPayload
	if err := json.Unmarshal(prepared.Payload, &preparedPayload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if preparedPayload.Template.Name != "Daily note for 2026-08-13" {
		t.Fatalf("template name = %q, want rendered value", preparedPayload.Template.Name)
	}
	content, ok := preparedPayload.Template.Blocks[0].ArborizedEditableBlock.Content.(cblocknote.InlineContentList)
	if !ok || len(content) != 1 {
		t.Fatalf("template block content = %#v, want one inline text item", preparedPayload.Template.Blocks[0].ArborizedEditableBlock.Content)
	}
	styledText, ok := content[0].InlineContentUnion.(*cblocknote.StyledText)
	if !ok || styledText.Text != "Daily note for 2026-08-13" {
		var got string
		if ok {
			got = styledText.Text
		}
		t.Fatalf("template block content = %q, want rendered value", got)
	}
	props, ok := preparedPayload.Template.Blocks[0].ArborizedEditableBlock.Props.(*cblocknote.BaseProps)
	if !ok || props.Template {
		t.Fatal("template marker should not be persisted in the prepared payload")
	}
}
