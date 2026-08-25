package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type DeleteMeRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			AuthCode string `json:"authCode" validate:"omitempty,isnumberstring,len=6"`
		},
		struct{},
		struct{},
	]
}

type DeleteMeResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
