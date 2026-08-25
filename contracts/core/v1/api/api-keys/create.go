package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type CreateMyAPIKeyRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Name      string     `json:"name" validate:"required,max=64"`
			ExpiresAt *time.Time `json:"expiresAt"`
		},
		struct{},
		struct{},
	]
}

type CreateMyAPIKeyResponseDto struct {
	PublicId  string     `json:"publicId"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"keyPrefix"`
	Secret    string     `json:"secret"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}
