package inputs

import (
	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateRoutineTagInput struct {
	Id    *uuid.UUID           `json:"id" gorm:"column:id;"`
	Name  string               `json:"name" gorm:"column:name;"`
	Color string               `json:"color" gorm:"column:color;"`
	Icon  *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
}

type UpdateRoutineTagInput struct {
	Name  *string              `json:"name" gorm:"column:name;"`
	Color *string              `json:"color" gorm:"column:color;"`
	Icon  *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
}

type PartialUpdateRoutineTagInput = PartialUpdateInput[UpdateRoutineTagInput]

type UpdateRoutineTagByIdInput struct {
	Id                 uuid.UUID                                 `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineTagInput] `json:"partialUpdateInput"`
}
