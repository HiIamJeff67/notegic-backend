package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type BindGoogleAccountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			AuthorizationCode string `json:"authorizationCode" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type BindGoogleAccountResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type UnbindGoogleAccountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			AuthCode string `json:"authCode" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type UnbindGoogleAccountResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
