package handlers

import (
	"gorm.io/datatypes"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
)

type RoutineTaskHandlerInterface interface {
	HandleValidateRoutineTaskPayload(
		purpose enums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *exceptions.Exception
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
) *exceptions.Exception {
	return h.payloadParser.ValidateRoutineTaskPayload(purpose, payload)
}
