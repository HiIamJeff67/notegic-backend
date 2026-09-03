package preparers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	durablejobexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/exceptions"
	validation "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/validations"
)

func TestPreparerPreparesAssignmentWithoutDatabaseAccess(t *testing.T) {
	payload, err := json.Marshal(croutinetasktypes.CreateMaterialRoutineTaskPayload{
		ParentSubShelfId: croutinetasktypes.RoutineTaskObjectReference(uuid.NewString()),
		Name:             "Daily {{date}}",
		ContentKey:       "material-key",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	assignment := croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineRecordId:     uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             cenums.RoutineTaskPurpose_CreateMaterial,
		Payload:             payload,
		Attempt:             1,
		ScheduledAt:         time.Now().UTC(),
		StartedAt:           time.Now().UTC(),
		PatternValues:       map[string]string{"date": "2026-08-05"},
	}

	prepared, exception := NewAssignmentPreparer(validation.New(), durablejobexceptions.NewRoutineTaskException()).Prepare(t.Context(), assignment)
	if exception != nil {
		t.Fatalf("prepare assignment: %v", exception)
	}
	if prepared == nil || prepared.RoutineTaskId != assignment.RoutineTaskId {
		t.Fatalf("prepared task = %#v", prepared)
	}

	var preparedPayload croutinetasktypes.CreateMaterialRoutineTaskPayload
	if err := json.Unmarshal(prepared.Payload, &preparedPayload); err != nil {
		t.Fatalf("decode prepared payload: %v", err)
	}
	if preparedPayload.Name != "Daily 2026-08-05" {
		t.Fatalf("prepared name = %q", preparedPayload.Name)
	}
}

func TestPreparerReturnsLocalErrorForInvalidPayload(t *testing.T) {
	assignment := croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineRecordId:     uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             cenums.RoutineTaskPurpose_GetMaterial,
		Payload:             []byte("{"),
	}

	prepared, err := NewAssignmentPreparer(validation.New(), durablejobexceptions.NewRoutineTaskException()).Prepare(t.Context(), assignment)
	if prepared != nil {
		t.Fatalf("prepared task = %#v, want nil", prepared)
	}
	if durableJobError, ok := err.(*cexceptions.Exception); !ok {
		t.Fatalf("error type = %T, want *exceptions.Exception", err)
	} else if durableJobError.Reason != "InvalidRoutineTaskPayload" || durableJobError.Domain != "RoutineTask" {
		t.Fatalf("error = %#v, want InvalidRoutineTaskPayload/RoutineTask", durableJobError)
	}
}

func TestPreparerRejectsRetiredPurpose(t *testing.T) {
	assignment := croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             cenums.RoutineTaskPurpose("AppendBlock"),
		Payload:             []byte(`{}`),
	}

	prepared, err := NewAssignmentPreparer(validation.New(), durablejobexceptions.NewRoutineTaskException()).Prepare(t.Context(), assignment)
	if prepared != nil {
		t.Fatalf("prepared task = %#v, want nil", prepared)
	}
	if err == nil {
		t.Fatal("retired AppendBlock purpose should be rejected")
	}
}

func TestPrepareAssignmentMatchesNestedTemplateBlockContent(t *testing.T) {
	payload, err := json.Marshal(croutinetasktypes.CreateBlockPackRoutineTaskPayload{
		TargetSubShelfId: croutinetasktypes.RoutineTaskObjectReference(uuid.NewString()),
		Template: croutinetasktypes.CreateBlockPackRoutineTaskTemplate{
			Name: "Daily note for {{date1}}",
			Blocks: []croutinetasktypes.CreateBlockPackRoutineTaskTemplateBlock{
				{
					ClientId: uuid.NewString(),
					ArborizedEditableBlock: cblocknote.ArborizedEditableBlock{
						Id:   uuid.New(),
						Type: cenums.BlockType_Paragraph,
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

	prepared, err := NewAssignmentPreparer(nil, durablejobexceptions.NewRoutineTaskException()).Prepare(nil, croutinetasktypes.RoutineTaskAssignment{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineRecordId:     uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserId:         uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             cenums.RoutineTaskPurpose_CreateBlockPack,
		Payload:             payload,
		PatternValues:       map[string]string{"date1": "2026-08-13"},
	})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
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
