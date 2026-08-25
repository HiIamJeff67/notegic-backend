package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
			Limit         int       `json:"limit" validate:"omitempty,min=1,max=500"`
		},
		struct{},
	]
}

type RoutineTaskRecordResponseDto struct {
	Id              uuid.UUID  `json:"id"`
	RoutineTaskId   uuid.UUID  `json:"routineTaskId"`
	Purpose         string     `json:"purpose"`
	Status          string     `json:"status"`
	ErrorCode       *string    `json:"errorCode"`
	ErrorReason     *string    `json:"errorReason"`
	CostUnit        int64      `json:"costUnit"`
	TotalAttempts   int64      `json:"totalAttempts"`
	ScheduledAt     time.Time  `json:"scheduledAt"`
	ActualStartedAt *time.Time `json:"actualStartedAt"`
	ActualEndedAt   *time.Time `json:"actualEndedAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type GetAllMyRoutineTaskRecordsByRoutineTaskIdResponseDto []RoutineTaskRecordResponseDto
