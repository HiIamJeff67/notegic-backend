package parsers

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	validation "github.com/HiIamJeff67/notegic-backend/runtimes/core/validations"
)

func TestValidateCreateBlockPackRoutineTaskPayloadAcceptsNestedBlockTree(t *testing.T) {
	parser := NewRoutineTaskPayloadParser(validation.New())
	payload := datatypes.JSON(`{
		"targetSubShelfId": "36cdc6db-ed4c-4f2a-a9b5-ed20401dfd4f",
		"template": {
			"name": "Daily note",
			"icon": null,
			"headerBackgroundURL": null,
			"blocks": [{
				"clientId": "842b2781-60c8-47a6-adb2-461d251ce04d",
				"prevClientId": null,
				"arborizedEditableBlock": {
					"id": "842b2781-60c8-47a6-adb2-461d251ce04d",
					"type": "bulletListItem",
					"props": {
						"backgroundColor": "default",
						"textColor": "default",
						"textAlignment": "left"
					},
					"content": [{
						"type": "text",
						"text": "Todo:",
						"styles": {}
					}],
					"children": [{
						"id": "b2fd031d-a2f7-43fb-9e08-fa51cb9f88c8",
						"type": "checkListItem",
						"props": {
							"backgroundColor": "default",
							"textColor": "default",
							"textAlignment": "left",
							"checked": false
						},
						"content": [{
							"type": "text",
							"text": "View documentation",
							"styles": {}
						}],
						"children": []
					}]
				}
			}]
		},
		"pattern": {}
	}`)

	if exception := parser.ValidateRoutineTaskPayload(
		cenums.RoutineTaskPurpose_CreateBlockPack,
		payload,
	); exception != nil {
		t.Fatalf("ValidateRoutineTaskPayload() exception = %v, want nil", exception)
	}
}

func TestValidateRoutineTaskPayloadAcceptsCurrentCrudPurposes(t *testing.T) {
	parser := NewRoutineTaskPayloadParser(validation.New())
	ids := map[string]uuid.UUID{
		"subShelfId":  uuid.New(),
		"blockPackId": uuid.New(),
		"routineId":   uuid.New(),
		"materialId":  uuid.New(),
	}
	payloads := map[cenums.RoutineTaskPurpose]datatypes.JSON{
		cenums.RoutineTaskPurpose_GetSubShelf:     datatypes.JSON(`{"subShelfId":"` + ids["subShelfId"].String() + `"}`),
		cenums.RoutineTaskPurpose_DeleteSubShelf:  datatypes.JSON(`{"subShelfId":"` + ids["subShelfId"].String() + `"}`),
		cenums.RoutineTaskPurpose_GetBlockPack:    datatypes.JSON(`{"blockPackId":"` + ids["blockPackId"].String() + `"}`),
		cenums.RoutineTaskPurpose_DeleteBlockPack: datatypes.JSON(`{"blockPackId":"` + ids["blockPackId"].String() + `"}`),
		cenums.RoutineTaskPurpose_GetRoutine:      datatypes.JSON(`{"routineId":"` + ids["routineId"].String() + `"}`),
		cenums.RoutineTaskPurpose_DeleteRoutine:   datatypes.JSON(`{"routineId":"` + ids["routineId"].String() + `"}`),
		cenums.RoutineTaskPurpose_GetMaterial:     datatypes.JSON(`{"materialId":"` + ids["materialId"].String() + `"}`),
		cenums.RoutineTaskPurpose_CreateMaterial:  datatypes.JSON(`{"parentSubShelfId":"` + ids["subShelfId"].String() + `","name":"note","contentKey":"key"}`),
		cenums.RoutineTaskPurpose_UpdateMaterial:  datatypes.JSON(`{"materialId":"` + ids["materialId"].String() + `","name":"updated"}`),
		cenums.RoutineTaskPurpose_DeleteMaterial:  datatypes.JSON(`{"materialId":"` + ids["materialId"].String() + `"}`),
	}

	for purpose, payload := range payloads {
		if exception := parser.ValidateRoutineTaskPayload(purpose, payload); exception != nil {
			t.Errorf("purpose %s returned exception: %v", purpose, exception)
		}
	}
}

func TestValidateRoutineTaskPayloadRejectsRetiredPurpose(t *testing.T) {
	parser := NewRoutineTaskPayloadParser(validation.New())
	if exception := parser.ValidateRoutineTaskPayload(
		cenums.RoutineTaskPurpose("AppendBlock"),
		datatypes.JSON(`{}`),
	); exception == nil {
		t.Fatal("retired AppendBlock purpose should be rejected")
	}
}

func TestValidateUpdateBlockPackRejectsMismatchedAndDuplicateBlockIds(t *testing.T) {
	parser := NewRoutineTaskPayloadParser(validation.New())
	blockPackId := uuid.New()
	blockId := uuid.New()
	blockPayload := func(id uuid.UUID) string {
		return `{"blockPackId":"` + blockPackId.String() + `","blocks":[{"blockId":"` + blockId.String() + `","arborizedEditableBlock":{"id":"` + id.String() + `","type":"paragraph","props":{},"content":[]}}]}`
	}

	if exception := parser.ValidateRoutineTaskPayload(
		cenums.RoutineTaskPurpose_UpdateBlockPack,
		datatypes.JSON(blockPayload(uuid.New())),
	); exception == nil {
		t.Fatal("mismatched block IDs should be rejected")
	}

	duplicatePayload := `{"blockPackId":"` + blockPackId.String() + `","blocks":[{"blockId":"` + blockId.String() + `","arborizedEditableBlock":{"id":"` + blockId.String() + `","type":"paragraph","props":{},"content":[]}},{"blockId":"` + blockId.String() + `","arborizedEditableBlock":{"id":"` + blockId.String() + `","type":"paragraph","props":{},"content":[]}}]}`
	if exception := parser.ValidateRoutineTaskPayload(
		cenums.RoutineTaskPurpose_UpdateBlockPack,
		datatypes.JSON(duplicatePayload),
	); exception == nil {
		t.Fatal("duplicate block IDs should be rejected")
	}
}

func TestValidateUpdateMaterialRequiresAField(t *testing.T) {
	parser := NewRoutineTaskPayloadParser(validation.New())
	payload, err := json.Marshal(croutinetasktypes.UpdateMaterialRoutineTaskPayload{MaterialId: uuid.New()})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if exception := parser.ValidateRoutineTaskPayload(cenums.RoutineTaskPurpose_UpdateMaterial, payload); exception == nil {
		t.Fatal("empty material update should be rejected")
	}
}
