package apicontract

import (
	"encoding/json"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
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
type VisualizeMyRoutineTaskPurposeCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			Permission cenums.AccessControlPermission `json:"permission" validate:"isaccesscontrolpermission,required"`
		},
		struct{},
	]
}
type VisualizeMyRoutineTaskPurposeCountResponseDto = RoutineTaskCountResponseDto
