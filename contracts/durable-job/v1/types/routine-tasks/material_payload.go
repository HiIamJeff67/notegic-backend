package routinetasktypes

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type GetMaterialRoutineTaskPayload struct {
	MaterialId uuid.UUID `json:"materialId" validate:"required"`
}

type CreateMaterialRoutineTaskPayload struct {
	Id               *uuid.UUID                  `json:"id" validate:"omitnil"`
	ParentSubShelfId RoutineTaskObjectReference  `json:"parentSubShelfId" validate:"required"`
	Name             string                      `json:"name" validate:"required,min=1,max=128"`
	ContentKey       string                      `json:"contentKey" validate:"required,min=1,max=512"`
	ContentType      *cenums.MaterialContentType `json:"contentType" validate:"omitnil,ismaterialcontenttype"`
	ParseMediaType   string                      `json:"parseMediaType" validate:"max=128"`
	Pattern          RoutineTaskPattern          `json:"pattern" validate:"omitempty,dive"`
}

type UpdateMaterialRoutineTaskPayload struct {
	MaterialId     uuid.UUID                   `json:"materialId" validate:"required"`
	Name           *string                     `json:"name" validate:"omitnil,min=1,max=128"`
	Size           *int64                      `json:"size" validate:"omitnil,gte=0"`
	ContentKey     *string                     `json:"contentKey" validate:"omitnil,min=1,max=512"`
	ContentType    *cenums.MaterialContentType `json:"contentType" validate:"omitnil,ismaterialcontenttype"`
	ParseMediaType *string                     `json:"parseMediaType" validate:"omitnil,max=128"`
	Pattern        RoutineTaskPattern          `json:"pattern" validate:"omitempty,dive"`
}

type DeleteMaterialRoutineTaskPayload struct {
	MaterialId uuid.UUID `json:"materialId" validate:"required"`
}
