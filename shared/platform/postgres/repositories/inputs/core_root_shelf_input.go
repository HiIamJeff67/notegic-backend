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

type UpdateRootShelfByIdInput struct {
	Id                 uuid.UUID                                `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRootShelfInput] `json:"partialUpdateInput"`
}
