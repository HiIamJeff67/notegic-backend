package apicontract

import (
	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RevokeMyAPIKeyRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			PublicId string `json:"publicId" validate:"required,uuid4"`
		},
		struct{},
	]
}

type RevokeMyAPIKeyResponseDto struct {
	RevokedAt string `json:"revokedAt"`
}
