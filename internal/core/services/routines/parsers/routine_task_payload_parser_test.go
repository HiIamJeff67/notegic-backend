package parsers

import (
	"testing"

	"gorm.io/datatypes"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	validation "github.com/HiIamJeff67/notegic-backend/internal/core/validations"
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
		enums.RoutineTaskPurpose_CreateBlockPack,
		payload,
	); exception != nil {
		t.Fatalf("ValidateRoutineTaskPayload() exception = %v, want nil", exception)
	}
}
