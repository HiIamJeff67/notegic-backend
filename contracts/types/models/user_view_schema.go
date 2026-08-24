package models

import (
	"github.com/google/uuid"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

// UserView is the minimal user projection that runtimes may read directly.
// Account and authentication details remain owned by the user runtime.
type UserView struct {
	PublicId uuid.UUID               `json:"publicId" gorm:"column:public_id;primaryKey;"`
	Status   enumcontract.UserStatus `json:"status" gorm:"column:status;type:\"UserStatus\";not null;"`
}

func (UserView) TableName() string {
	return "UserView"
}
