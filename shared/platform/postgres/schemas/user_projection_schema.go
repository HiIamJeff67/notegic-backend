package schemas

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// UserProjection is the Notification runtime's local user snapshot. It is
// intentionally limited to fields needed to validate notification recipients.
type UserProjection struct {
	Id       uuid.UUID         `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid();"`
	PublicId uuid.UUID         `json:"publicId" gorm:"column:public_id;type:uuid;not null;uniqueIndex;"`
	Plan     cenums.UserPlan   `json:"plan" gorm:"column:plan;type:\"UserPlan\";not null;"`
	Status   cenums.UserStatus `json:"status" gorm:"column:status;type:\"UserStatus\";not null;"`
}

func (UserProjection) TableName() string {
	return postgres.TableName_UserProjection.String()
}
