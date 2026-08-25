package inputs

import (
	"time"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdateUserDataCacheInput struct {
	DisplayName *string
	Email       *string
	AccessToken *string
	CSRFToken   *string
	Role        *enums.UserRole
	Plan        *enums.UserPlan
	Status      *enums.UserStatus
	AvatarURL   *string
}

type CheckAndUpdateUserQuotaInput struct {
	Field        string
	ChangeAmount int32
	MaxLimit     int32
	ExpiresIn    time.Time
}

type BatchCheckAndUpdateUserQuotaInput struct {
	Identifier string
	Input      CheckAndUpdateUserQuotaInput
}
