package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type UpdateMyAccountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			AuthCode string `json:"authCode" validate:"required,isnumberstring,len=6"`
			Values   struct {
				CountryCode *string `json:"countryCode" validate:"omitnil,iscountrycode"`
				BackupEmail *string `json:"backupEmail" validate:"omitnil,email"`
				PhoneNumber *string `json:"phoneNumber" validate:"omitnil,min=1,max=15,isnumberstring"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct{},
		struct{},
	]
}

type UpdateMyAccountResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
