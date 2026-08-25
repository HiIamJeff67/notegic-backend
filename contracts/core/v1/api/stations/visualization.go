package apicontract

import coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"

type VisualizeMyTotalCountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct {
			Permission string `json:"permission" validate:"required,isaccesscontrolpermission"`
		},
	]
}

type TotalCountDatumResponseDto struct {
	Id    string `json:"id"`
	X     string `json:"x"`
	Value int64  `json:"value"`
}

type VisualizeMyTotalCountResponseDto struct {
	Data []TotalCountDatumResponseDto `json:"data"`
}
