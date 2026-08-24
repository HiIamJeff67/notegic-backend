package apicontract

import (
	"encoding/json"
	"time"

	coreapicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineCountDatum struct {
	Id    string          `json:"id"`
	X     string          `json:"x"`
	Value int64           `json:"value"`
	Meta  json.RawMessage `json:"meta"`
}
type RoutineCountResponseDto struct {
	Data []RoutineCountDatum `json:"data"`
}
type VisualizeMyRoutineStatusCountRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission enumcontract.AccessControlPermission `json:"permission" validate:"isaccesscontrolpermission,required"`
		},
		struct{},
	]
}
type VisualizeMyRoutineStatusCountResponseDto = RoutineCountResponseDto
type VisualizeMyRoutinePeriodCountRequestDto = VisualizeMyRoutineStatusCountRequestDto
type VisualizeMyRoutinePeriodCountResponseDto = RoutineCountResponseDto
type VisualizeMyRoutineScheduledStartAtCountRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission          enumcontract.AccessControlPermission `json:"permission" validate:"isaccesscontrolpermission,required"`
			TimeHourUnit        int                                  `json:"timeHourUnit" validate:"required,min=1"`
			QueryRangeStartedAt time.Time                            `json:"queryRangeStartedAt" validate:"required"`
			QueryRangeEndedAt   time.Time                            `json:"queryRangeEndedAt" validate:"required"`
		},
		struct{},
	]
}
type VisualizeMyRoutineScheduledStartAtCountResponseDto = RoutineCountResponseDto
type VisualizeMyRoutineScheduledEndAtCountRequestDto = VisualizeMyRoutineScheduledStartAtCountRequestDto
type VisualizeMyRoutineScheduledEndAtCountResponseDto = RoutineCountResponseDto
