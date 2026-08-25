package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type ForgetPasswordRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Account     string `json:"account" validate:"required,isaccount"`
			NewPassword string `json:"newPassword" validate:"required,min=8,max=1024,isstrongpassword"`
			AuthCode    string `json:"authCode" validate:"required,isnumberstring,len=6"`
		},
		struct{},
		struct{},
	]
}

type ForgetPasswordResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
