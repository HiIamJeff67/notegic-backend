package inputs

import (
	"github.com/google/uuid"
)

/* ============================== System Only Input ============================== */

type BulkCheckMaterialPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkDeleteMaterialInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}
