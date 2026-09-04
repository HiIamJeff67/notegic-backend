package apicontract

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskResponseDto struct {
	Id                     uuid.UUID                 `json:"id"`
	RoutineId              uuid.UUID                 `json:"routineId"`
	Title                  string                    `json:"title"`
	Purpose                cenums.RoutineTaskPurpose `json:"purpose"`
	Phase                  *cenums.RoutinePhase      `json:"phase"`
	Payload                datatypes.JSON            `json:"payload"`
	CostUnit               int64                     `json:"costUnit"`
	Priority               int32                     `json:"priority"`
	MaxAttempts            int32                     `json:"maxAttempts"`
	PreviousRoutineTaskIds []uuid.UUID               `json:"previousRoutineTaskIds"`
	UpdatedAt              time.Time                 `json:"updatedAt"`
	CreatedAt              time.Time                 `json:"createdAt"`
}

type GetMyRoutineTaskByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
			IsDeleted     *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutineTaskByIdResponseDto = RoutineTaskResponseDto

type GetMyRoutineTasksByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineId  uuid.UUID `json:"routineId" validate:"required"`
			AreDeleted *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMyRoutineTasksByRoutineIdResponseDto []RoutineTaskResponseDto

type GetMyRoutineTasksByRoutineIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineIds []uuid.UUID `json:"routineIds" validate:"required,min=1,max=1024"`
			AreDeleted *bool       `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutineTasksByRoutineIdsResponseDto []RoutineTaskResponseDto
type GetAllMyRoutineTasksRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			AreDeleted *bool `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetAllMyRoutineTasksResponseDto []RoutineTaskResponseDto
