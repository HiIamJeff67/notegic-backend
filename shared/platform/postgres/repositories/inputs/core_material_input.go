package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateMaterialInput struct {
	Id             uuid.UUID                  `json:"id" gorm:"column:id;"`
	Name           string                     `json:"name" gorm:"column:name;"`
	Size           int64                      `json:"size" gorm:"column:size;"`
	ContentKey     string                     `json:"contentKey" gorm:"column:content_key;"`
	ContentType    cenums.MaterialContentType `json:"contentType" gorm:"column:content_type;"`
	ParseMediaType string                     `json:"parseMediaType" gorm:"column:parse_media_type;"`
}

type UpdateMaterialInput struct {
	ParentSubShelfId *uuid.UUID                  `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id;"`
	Name             *string                     `json:"name" gorm:"column:name;"`
	Size             *int64                      `json:"size" gorm:"column:size;"`
	ContentKey       *string                     `json:"contentKey" gorm:"column:content_key;"`
	ContentType      *cenums.MaterialContentType `json:"contentType" gorm:"column:content_type;"`
	ParseMediaType   string                      `json:"parseMediaType" gorm:"column:parse_media_type;"`
}

type PartialUpdateMaterialInput = PartialUpdateInput[UpdateMaterialInput]
