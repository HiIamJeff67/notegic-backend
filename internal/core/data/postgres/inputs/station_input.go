package inputs

import (
	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type CreateStationInput struct {
	Id                  *uuid.UUID           `json:"id" gorm:"column:id;"`
	Name                string               `json:"name" gorm:"column:name;"`
	Description         string               `json:"description" gorm:"column:description;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"column:header_background_url;"`
}

type UpdateStationInput struct {
	Name                *string              `json:"name" gorm:"column:name;"`
	Description         *string              `json:"description" gorm:"column:description;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"column:header_background_url;"`
}

type PartialUpdateStationInput = PartialUpdateInput[UpdateStationInput]

type UpdateStationByIdInput struct {
	Id                 uuid.UUID                              `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateStationInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckStationPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateStationInput struct {
	UserId              uuid.UUID            `json:"userId" gorm:"column:user_id;"`
	Id                  *uuid.UUID           `json:"id" gorm:"column:id;"`
	Name                string               `json:"name" gorm:"column:name;"`
	Description         string               `json:"description" gorm:"column:description;"`
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"column:header_background_url;"`
}

type BulkUpdateStationInput struct {
	UserId             uuid.UUID                              `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                              `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateStationInput] `json:"partialUpdateInput"`
}
