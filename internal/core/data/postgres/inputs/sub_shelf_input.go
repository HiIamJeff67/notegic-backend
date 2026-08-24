package inputs

import "github.com/google/uuid"

type CreateSubShelfInput struct {
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	PrevSubShelfId *uuid.UUID `json:"prevSubShelfId" gorm:"column:prev_sub_shelf_id;"`
	Name           string     `json:"name" gorm:"column:name;"`
	// will be automatically set to the path of the prevSubShelf
	// Path []uuid.UUID `json:"path" gorm:"column:path;"`
}

type CreateSubShelfByRootShelfIdInput struct {
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	RootShelfId    uuid.UUID  `json:"rootShelfId" gorm:"column:root_shelf_id;"`
	PrevSubShelfId *uuid.UUID `json:"prevSubShelfId" gorm:"column:prev_sub_shelf_id;"`
	Name           string     `json:"name" gorm:"column:name;"`
}

type UpdateSubShelfInput struct {
	Name *string `json:"name" gorm:"column:name;"`
}

type PartialUpdateSubShelfInput = PartialUpdateInput[UpdateSubShelfInput]

type UpdateSubShelfByIdInput struct {
	Id                 uuid.UUID                               `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateSubShelfInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckSubShelfPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateSubShelfInput struct {
	UserId         uuid.UUID  `json:"userId" gorm:"column:user_id;"`
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	RootShelfId    uuid.UUID  `json:"rootShelfId" gorm:"column:root_shelf_id;"`
	PrevSubShelfId *uuid.UUID `json:"prevSubShelfId" gorm:"column:prev_sub_shelf_id;"`
	Name           string     `json:"name" gorm:"column:name;"`
}

type BulkUpdateSubShelfInput struct {
	UserId             uuid.UUID                               `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                               `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateSubShelfInput] `json:"partialUpdateInput"`
}

type BulkDeleteSubShelfInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}
