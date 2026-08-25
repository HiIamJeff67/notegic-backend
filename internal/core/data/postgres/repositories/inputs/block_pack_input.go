package inputs

import (
	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateBlockPackInput struct {
	Id                  *uuid.UUID           `json:"id" gorm:"column:id;"`
	Name                string               `json:"name" gorm:"column:name;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type CreateBlockPackBySubShelfIdInput struct {
	Id                  *uuid.UUID           `json:"id" gorm:"column:id;"`
	ParentSubShelfId    uuid.UUID            `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                string               `json:"name" gorm:"column:name;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type UpdateBlockPackInput struct {
	ParentSubShelfId    *uuid.UUID           `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                *string              `json:"name" gorm:"column:name;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"header_background_url;"`
}

type PartialUpdateBlockPackInput = PartialUpdateInput[UpdateBlockPackInput]

type UpdateBlockPackByIdInput struct {
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateBlockPackInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckBlockPackPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateBlockPackInput struct {
	UserId              uuid.UUID            `json:"userId" gorm:"column:user_id;"`
	Id                  *uuid.UUID           `json:"id" gorm:"column:id;"`
	ParentSubShelfId    uuid.UUID            `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name                string               `json:"name" gorm:"column:name;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"header_background_url;"`
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
