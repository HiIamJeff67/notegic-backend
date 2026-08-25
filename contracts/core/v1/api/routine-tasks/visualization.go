package apicontract

import (
	"encoding/json"
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskCountDatum struct {
	Id    string          `json:"id"`
	X     string          `json:"x"`
	Value int64           `json:"value"`
	Meta  json.RawMessage `json:"meta"`
}
type RoutineTaskCountResponseDto struct {
	Data []RoutineTaskCountDatum `json:"data"`
}
type VisualizeMyRoutineTaskStatusCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission enums.AccessControlPermission `json:"permission" validate:"isaccesscontrolpermission,required"`
		},
		struct{},
	]
}
type VisualizeMyRoutineTaskStatusCountResponseDto = RoutineTaskCountResponseDto
type VisualizeMyRoutineTaskPurposeCountRequestDto = VisualizeMyRoutineTaskStatusCountRequestDto
type VisualizeMyRoutineTaskPurposeCountResponseDto = RoutineTaskCountResponseDto
type VisualizeMyRoutineTaskScheduledAtCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission          enums.AccessControlPermission `json:"permission" validate:"isaccesscontrolpermission,required"`
			TimeHourUnit        int                           `json:"timeHourUnit" validate:"required,min=1"`
			QueryRangeStartedAt time.Time                     `json:"queryRangeStartedAt" validate:"required"`
			QueryRangeEndedAt   time.Time                     `json:"queryRangeEndedAt" validate:"required"`
		},
		struct{},
	]
}
type VisualizeMyRoutineTaskScheduledAtCountResponseDto = RoutineTaskCountResponseDto
type VisualizeMyRoutineTaskActualStartedAtCountRequestDto = VisualizeMyRoutineTaskScheduledAtCountRequestDto
type VisualizeMyRoutineTaskActualStartedAtCountResponseDto = RoutineTaskCountResponseDto
type VisualizeMyRoutineTaskActualEndedAtCountRequestDto = VisualizeMyRoutineTaskScheduledAtCountRequestDto
type VisualizeMyRoutineTaskActualEndedAtCountResponseDto = RoutineTaskCountResponseDto
