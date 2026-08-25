package schemas

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// UserView is the minimal user projection that runtimes may read directly.
// Account and authentication details remain owned by the user runtime.
type UserView struct {
	Id        uuid.UUID        `json:"id" gorm:"column:id;primaryKey;"`
	PublicId  uuid.UUID        `json:"publicId" gorm:"column:public_id;"`
	Plan      enums.UserPlan   `json:"plan" gorm:"column:plan;type:\"UserPlan\";not null;"`
	Status    enums.UserStatus `json:"status" gorm:"column:status;type:\"UserStatus\";not null;"`
	CreatedAt time.Time        `json:"createdAt" gorm:"column:created_at;not null;"`
}

func (UserView) TableName() string {
	return platformpostgres.TableName_UserView.String()
}
