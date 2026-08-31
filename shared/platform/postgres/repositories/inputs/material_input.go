package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type BulkCreateMaterialInput struct {
	UserId           uuid.UUID                  `json:"userId" gorm:"column:user_id;"`
	Id               *uuid.UUID                 `json:"id" gorm:"column:id;"`
	ParentSubShelfId uuid.UUID                  `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name             string                     `json:"name" gorm:"column:name;"`
	Size             int64                      `json:"size" gorm:"column:size;"`
	ContentKey       string                     `json:"contentKey" gorm:"column:content_key;"`
	ContentType      cenums.MaterialContentType `json:"contentType" gorm:"column:content_type;"`
	ParseMediaType   string                     `json:"parseMediaType" gorm:"column:parse_media_type;"`
}

type BulkUpdateMaterialInput struct {
	UserId             uuid.UUID                               `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                               `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateMaterialInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckMaterialPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkDeleteMaterialInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}
