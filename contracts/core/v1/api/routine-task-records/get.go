package apicontract

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyRoutineTaskRecordsByRoutineTaskIdRequestDto struct {
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
	Id              uuid.UUID      `json:"id"`
	RoutineRecordId uuid.UUID      `json:"routineRecordId"`
	RoutineTaskId   uuid.UUID      `json:"routineTaskId"`
	Purpose         string         `json:"purpose"`
	Status          string         `json:"status"`
	ErrorCode       *string        `json:"errorCode"`
	ErrorReason     *string        `json:"errorReason"`
	CostUnit        int64          `json:"costUnit"`
	Attempts        int32          `json:"attempts"`
	PayloadSnapshot datatypes.JSON `json:"payloadSnapshot"`
	ResultSnapshot  datatypes.JSON `json:"resultSnapshot"`
	ActualStartedAt *time.Time     `json:"actualStartedAt"`
	ActualEndedAt   *time.Time     `json:"actualEndedAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type GetMyRoutineTaskRecordsByRoutineTaskIdResponseDto []RoutineTaskRecordResponseDto
