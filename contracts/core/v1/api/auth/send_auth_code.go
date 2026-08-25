package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type SendAuthCodeRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Email string `json:"email" validate:"required,email"`
		},
		struct{},
		struct{},
	]
}

type SendAuthCodeResponseDto struct {
	AuthCodeExpiredAt  time.Time `json:"authCodeExpiredAt"`
	BlockAuthCodeUntil time.Time `json:"blockAuthCodeUntil"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
