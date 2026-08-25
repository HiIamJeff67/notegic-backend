package inputs

import (
	"github.com/google/uuid"
)

type CreateSubShelfInput struct {
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	PrevSubShelfId *uuid.UUID `json:"prevSubShelfId" gorm:"column:prev_sub_shelf_id;"`
	Name           string     `json:"name" gorm:"column:name;"`
}

type CreateSubShelfByRootShelfIdInput struct {
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	RootShelfId    uuid.UUID  `json:"rootShelfId" gorm:"column:root_shelf_id;"`
	PrevSubShelfId *uuid.UUID `json:"prevSubShelfId" gorm:"column:prev_sub_shelf_id;"`
	Name           string     `json:"name" gorm:"column:name;"`
}

type UpdateSubShelfByIdInput struct {
	Id                 uuid.UUID                               `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateSubShelfInput] `json:"partialUpdateInput"`
}
