package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type LoginRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Account  string `json:"account" validate:"required,isaccount"`
			Password string `json:"password" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type LoginResponseDto struct {
	PublicId     uuid.UUID `json:"publicId"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"displayName"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	CSRFToken    string    `json:"csrfToken"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type LoginViaGoogleRequestDto struct {
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

type LoginViaGoogleResponseDto struct {
	PublicId     uuid.UUID `json:"publicId"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"displayName"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	CSRFToken    string    `json:"csrfToken"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}
