package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetUserDataRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type GetUserDataResponseDto struct {
	PublicId    uuid.UUID `json:"publicId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Plan        string    `json:"plan"`
	Status      string    `json:"status"`
	AvatarURL   string    `json:"avatarURL"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GetMeRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type GetMeResponseDto struct {
	PublicId    uuid.UUID `json:"publicId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Plan        string    `json:"plan"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
