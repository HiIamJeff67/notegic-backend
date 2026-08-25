package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type UpdateMyInfoRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values struct {
				CoverBackgroundURL *string    `json:"coverBackgroundURL,omitempty" validate:"omitnil,isimageurl"`
				AvatarURL          *string    `json:"avatarURL,omitempty" validate:"omitnil,isimageurl"`
				Header             *string    `json:"header,omitempty" validate:"omitnil,min=0,max=64"`
				Introduction       *string    `json:"introduction,omitempty" validate:"omitnil,min=0,max=256"`
				Gender             *string    `json:"gender,omitempty" validate:"omitnil,isgender"`
				Country            *string    `json:"country,omitempty" validate:"omitnil,iscountry"`
				BirthDate          *time.Time `json:"birthDate,omitempty" validate:"omitnil,notfuture"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct{},
		struct{},
	]
}

type UpdateMyInfoResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
