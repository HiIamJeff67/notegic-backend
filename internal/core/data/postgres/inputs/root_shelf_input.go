package inputs

import (
	"time"

	"github.com/google/uuid"
)

type CreateRootShelfInput struct {
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	Name           string     `json:"name" gorm:"column:name;"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt" gorm:"column:last_analyzed_at;"`
}

type UpdateRootShelfInput struct {
	Name           *string    `json:"name" gorm:"column:name;"`
	SubShelfCount  *int64     `json:"subShelfCount" gorm:"column:sub_shelf_count;"`
	ItemCount      *int64     `json:"itemCount" gorm:"column:item_count;"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt" gorm:"column:last_analyzed_at;"`
}

type PartialUpdateRootShelfInput = PartialUpdateInput[UpdateRootShelfInput]

type UpdateRootShelfByIdInput struct {
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRootShelfInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckRootShelfPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateRootShelfInput struct {
	UserId         uuid.UUID  `json:"userId" gorm:"column:user_id;"`
	Id             *uuid.UUID `json:"id" gorm:"column:id;"`
	Name           string     `json:"name" gorm:"column:name;"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt" gorm:"column:last_analyzed_at;"`
}

type BulkUpdateRootShelfInput struct {
	UserId             uuid.UUID                                `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRootShelfInput] `json:"partialUpdateInput"`
}
