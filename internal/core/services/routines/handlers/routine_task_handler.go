package handlers

import (
	"gorm.io/datatypes"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
)

type RoutineTaskHandlerInterface interface {
	HandleValidateRoutineTaskPayload(
		purpose enums.RoutineTaskPurpose,
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
	purpose enums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *cexceptions.Exception {
	return h.payloadParser.ValidateRoutineTaskPayload(purpose, payload)
}
