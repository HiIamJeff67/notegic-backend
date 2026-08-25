package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	durablejobexceptions "github.com/HiIamJeff67/notegic-backend/internal/durablejob/exceptions"
)

type PurposeHandler struct {
	HandlerFunc PurposeHandlerFunc
}

type PurposeHandlerFunc func(
	context.Context,
	croutinetasktypes.RoutineTaskAssignment,
) (*croutinetasktypes.PreparedRoutineTask, error)

func NewPurposeHandler(validator *validator.Validate) PurposeHandler {
	return PurposeHandler{
		HandlerFunc: func(
			ctx context.Context,
			assignment croutinetasktypes.RoutineTaskAssignment,
		) (*croutinetasktypes.PreparedRoutineTask, error) {
			return prepareAssignment(ctx, validator, assignment)
		},
	}
}

func prepareAssignment(
	_ context.Context,
	validator *validator.Validate,
	assignment croutinetasktypes.RoutineTaskAssignment,
) (*croutinetasktypes.PreparedRoutineTask, error) {
	if assignment.RoutineTaskId == uuid.Nil || assignment.RoutineTaskRecordId == uuid.Nil ||
		assignment.RoutineId == uuid.Nil || assignment.ActorUserId == uuid.Nil || assignment.ActorUserPublicId == uuid.Nil ||
		assignment.Purpose == "" || len(assignment.Payload) == 0 {
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(
			fmt.Errorf("routine task assignment is incomplete"),
		)
	}

	var payload any
	switch assignment.Purpose {
	case enums.RoutineTaskPurpose_CreateRootShelf:
		payload = &croutinetasktypes.CreateRootShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_UpdateRootShelf:
		payload = &croutinetasktypes.UpdateRootShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_ResetRootShelf:
		payload = &croutinetasktypes.ResetRootShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_CreateSubShelf:
		payload = &croutinetasktypes.CreateSubShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_UpdateSubShelf:
		payload = &croutinetasktypes.UpdateSubShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_ResetSubShelf:
		payload = &croutinetasktypes.ResetSubShelfRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_CreateBlockPack:
		payload = &croutinetasktypes.CreateBlockPackRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_UpdateBlockPack:
		payload = &croutinetasktypes.UpdateBlockPackRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_ResetBlockPack:
		payload = &croutinetasktypes.ResetBlockPackRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_AppendBlock:
		payload = &croutinetasktypes.AppendBlockRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_UpdateBlock:
		payload = &croutinetasktypes.UpdateBlockRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_ResetBlock:
		payload = &croutinetasktypes.ResetBlockRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_CreateRoutine:
		payload = &croutinetasktypes.CreateRoutineRoutineTaskPayload{}
	case enums.RoutineTaskPurpose_UpdateRoutine:
		payload = &croutinetasktypes.UpdateRoutineRoutineTaskPayload{}
	default:
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(
			fmt.Errorf("unsupported routine task purpose: %s", assignment.Purpose),
		)
	}

	if err := json.Unmarshal(assignment.Payload, payload); err != nil {
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(err)
	}
	if validator != nil {
		if err := validator.Struct(payload); err != nil {
			return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(err)
		}
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(err)
	}
	var payloadValue any
	if err := json.Unmarshal(rawPayload, &payloadValue); err != nil {
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(err)
	}
	payloadValue = matchPayloadValue(payloadValue, assignment.PatternValues, true)
	preparedPayload, err := json.Marshal(payloadValue)
	if err != nil {
		return nil, durablejobexceptions.NewRoutineTaskException("RoutineTask").InvalidPayload(err)
	}

	return &croutinetasktypes.PreparedRoutineTask{
		RoutineTaskId:       assignment.RoutineTaskId,
		RoutineTaskRecordId: assignment.RoutineTaskRecordId,
		RoutineId:           assignment.RoutineId,
		ActorUserId:         assignment.ActorUserId,
		ActorUserPublicId:   assignment.ActorUserPublicId,
		Attempt:             assignment.Attempt,
		Purpose:             assignment.Purpose,
		Payload:             preparedPayload,
		PreparedAt:          time.Now().UTC(),
	}, nil
}

func matchPayloadValue(value any, values map[string]string, allowStrings bool) any {
	if len(values) == 0 {
		return value
	}

	switch typed := value.(type) {
	case string:
		if !allowStrings {
			return typed
		}
		matched := typed
		for key, resolvedValue := range values {
			matched = strings.ReplaceAll(matched, "{{"+key+"}}", resolvedValue)
		}
		return matched
	case []any:
		matched := make([]any, len(typed))
		for index, item := range typed {
			matched[index] = matchPayloadValue(item, values, allowStrings)
		}
		return matched
	case map[string]any:
		isTemplateBlock := false
		isNestedTemplateBlock := false
		if props, ok := typed["props"].(map[string]any); ok {
			if template, ok := props["template"].(bool); ok {
				isTemplateBlock = template
				if template {
					delete(props, "template")
				}
			}
		}
		if arborizedEditableBlock, ok := typed["arborizedEditableBlock"].(map[string]any); ok {
			if props, ok := arborizedEditableBlock["props"].(map[string]any); ok {
				if template, ok := props["template"].(bool); ok {
					isNestedTemplateBlock = template
					if template {
						delete(props, "template")
					}
				}
			}
		}

		matched := make(map[string]any, len(typed))
		for key, item := range typed {
			if key == "pattern" {
				continue
			}
			childAllowsStrings := allowStrings
			if key == "arborizedEditableBlock" {
				childAllowsStrings = isNestedTemplateBlock
			} else if key == "children" {
				childAllowsStrings = isTemplateBlock
			}
			matched[key] = matchPayloadValue(item, values, childAllowsStrings)
		}
		return matched
	default:
		return value
	}
}
