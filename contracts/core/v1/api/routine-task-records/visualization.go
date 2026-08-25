package apicontract

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RoutineTaskRecordCountDatum struct {
	Id    string          `json:"id"`
	X     string          `json:"x"`
	Value int64           `json:"value"`
	Meta  json.RawMessage `json:"meta"`
}

type RoutineTaskRecordCountResponseDto struct {
	Data []RoutineTaskRecordCountDatum `json:"data"`
}

type VisualizeMyRoutineTaskRecordStatusCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission     string      `json:"permission" validate:"required,isaccesscontrolpermission"`
			RoutineTaskIds []uuid.UUID `json:"routineTaskIds" validate:"omitempty,max=1024"`
		},
		struct{},
	]
}

type VisualizeMyRoutineTaskRecordStatusCountResponseDto = RoutineTaskRecordCountResponseDto

type VisualizeMyRoutineTaskRecordPurposeCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission     string      `json:"permission" validate:"required,isaccesscontrolpermission"`
			RoutineTaskIds []uuid.UUID `json:"routineTaskIds" validate:"omitempty,max=1024"`
		},
		struct{},
	]
}

type VisualizeMyRoutineTaskRecordPurposeCountResponseDto = RoutineTaskRecordCountResponseDto

type VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission          string      `json:"permission" validate:"required,isaccesscontrolpermission"`
			RoutineTaskIds      []uuid.UUID `json:"routineTaskIds" validate:"omitempty,max=1024"`
			TimeHourUnit        int         `json:"timeHourUnit" validate:"required,min=1"`
			QueryRangeStartedAt time.Time   `json:"queryRangeStartedAt" validate:"required"`
			QueryRangeEndedAt   time.Time   `json:"queryRangeEndedAt" validate:"required"`
		},
		struct{},
	]
}

type VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto = RoutineTaskRecordCountResponseDto

type VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission          string      `json:"permission" validate:"required,isaccesscontrolpermission"`
			RoutineTaskIds      []uuid.UUID `json:"routineTaskIds" validate:"omitempty,max=1024"`
			TimeHourUnit        int         `json:"timeHourUnit" validate:"required,min=1"`
			QueryRangeStartedAt time.Time   `json:"queryRangeStartedAt" validate:"required"`
			QueryRangeEndedAt   time.Time   `json:"queryRangeEndedAt" validate:"required"`
		},
		struct{},
	]
}

type VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto = RoutineTaskRecordCountResponseDto

type VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission          string      `json:"permission" validate:"required,isaccesscontrolpermission"`
			RoutineTaskIds      []uuid.UUID `json:"routineTaskIds" validate:"omitempty,max=1024"`
			TimeHourUnit        int         `json:"timeHourUnit" validate:"required,min=1"`
			QueryRangeStartedAt time.Time   `json:"queryRangeStartedAt" validate:"required"`
			QueryRangeEndedAt   time.Time   `json:"queryRangeEndedAt" validate:"required"`
		},
		struct{},
	]
}

type VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto = RoutineTaskRecordCountResponseDto
