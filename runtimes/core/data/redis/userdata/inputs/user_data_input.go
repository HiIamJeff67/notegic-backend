package inputs

import (
	"time"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdateUserDataCacheInput struct {
	DisplayName *string
	Email       *string
	AccessToken *string
	CSRFToken   *string
	Role        *cenums.UserRole
	Plan        *cenums.UserPlan
	Status      *cenums.UserStatus
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
