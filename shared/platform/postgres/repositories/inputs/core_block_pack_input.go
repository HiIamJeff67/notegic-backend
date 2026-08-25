package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateBlockPackInput struct {
	Id                  *uuid.UUID            `json:"id" gorm:"column:id;"`
	Name                string                `json:"name" gorm:"column:name;"`
	Icon                *cenums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type CreateBlockPackBySubShelfIdInput struct {
	Id                  *uuid.UUID            `json:"id" gorm:"column:id;"`
	ParentSubShelfId    uuid.UUID             `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                string                `json:"name" gorm:"column:name;"`
	Icon                *cenums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type PartialUpdateBlockPackInput = PartialUpdateInput[UpdateBlockPackInput]

type UpdateBlockPackByIdInput struct {
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateBlockPackInput] `json:"partialUpdateInput"`
}
