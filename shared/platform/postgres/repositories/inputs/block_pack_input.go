package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdateBlockPackInput struct {
	ParentSubShelfId    *uuid.UUID            `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                *string               `json:"name" gorm:"column:name;"`
	Icon                *cenums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

/* ============================== System Only Input ============================== */

type BulkCheckBlockPackPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateBlockPackInput struct {
	UserId              uuid.UUID             `json:"userId" gorm:"column:user_id;"`
	Id                  *uuid.UUID            `json:"id" gorm:"column:id;"`
	ParentSubShelfId    uuid.UUID             `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                string                `json:"name" gorm:"column:name;"`
	Icon                *cenums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type BulkUpdateBlockPackInput struct {
	UserId             uuid.UUID                                `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateBlockPackInput] `json:"partialUpdateInput"`
}

type BulkDeleteBlockPackInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}
