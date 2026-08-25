package handlers

import (
	"gorm.io/datatypes"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
)

type RoutineTaskHandlerInterface interface {
	HandleValidateRoutineTaskPayload(
		purpose cenums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *cexceptions.Exception
}

type RoutineTaskHandler struct {
	payloadParser parsers.RoutineTaskPayloadParserInterface
}

func NewRoutineTaskHandler(
	payloadParser parsers.RoutineTaskPayloadParserInterface,
) RoutineTaskHandlerInterface {
	return &RoutineTaskHandler{payloadParser: payloadParser}
}

func (h *RoutineTaskHandler) HandleValidateRoutineTaskPayload(
	purpose cenums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *cexceptions.Exception {
	return h.payloadParser.ValidateRoutineTaskPayload(purpose, payload)
}
