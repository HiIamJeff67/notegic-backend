package durablejobcontract

import (
	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
)

type ClaimRoutinesRequestDto struct {
	RequestId uuid.UUID `json:"requestId" validate:"required"`
	WorkerId  uuid.UUID `json:"workerId" validate:"required"`
	BatchSize int       `json:"batchSize" validate:"required,min=1,max=1000"`
}

type ClaimRoutinesResponseDto struct {
	RequestId          uuid.UUID                             `json:"requestId"`
	WorkerId           uuid.UUID                             `json:"workerId"`
	RoutineAssignments []croutinetasktypes.RoutineAssignment `json:"routineAssignments"`
}

type MarkCompletedRoutineTasksRequestDto struct {
	WorkerId uuid.UUID                                `json:"workerId" validate:"required"`
	Tasks    []croutinetasktypes.CompletedRoutineTask `json:"tasks" validate:"required,min=1,dive"`
}

type MarkFailedRoutineTasksRequestDto struct {
	WorkerId uuid.UUID                             `json:"workerId" validate:"required"`
	Tasks    []croutinetasktypes.FailedRoutineTask `json:"tasks" validate:"required,min=1,dive"`
}
