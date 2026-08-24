package schemas

import (
	"time"

	"github.com/google/uuid"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type Material struct {
	Id               uuid.UUID                 `json:"id" gorm:"column:id; type:uuid; primaryKey; not null;"`
	ParentSubShelfId uuid.UUID                 `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id; type:uuid; not null;"` // Previous unique-name constraint: uniqueIndex:material_idx_parent_sub_shelf_id_name,where:deleted_at IS NULL
	Name             string                    `json:"name" gorm:"column:name; size:128; not null; default:'undefined';"`        // Previous unique-name constraint: uniqueIndex:material_idx_parent_sub_shelf_id_name,where:deleted_at IS NULL
	Size             int64                     `json:"size" gorm:"column:size; type:bigint; not null; default:0;"`
	ContentKey       string                    `json:"contentKey" gorm:"column:content_key; unique; not null;"`
	ContentType      enums.MaterialContentType `json:"contentType" gorm:"column:content_type; type:\"MaterialContentType\"; not null; default:'none';"`
	ParseMediaType   string                    `json:"parseMediaType" gorm:"column:parse_media_type; size:128; not null; default:'';"`
	DeletedAt        *time.Time                `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt        time.Time                 `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt        time.Time                 `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	ParentSubShelf SubShelf `json:"parentSubShelf" gorm:"foreignKey:ParentSubShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Constraints:
// Duplicate names are allowed. Previously, Material.ParentSubShelfId + Material.Name was unique while DeletedAt was NULL.

// Material Table Name
func (Material) TableName() string {
	return "MaterialTable"
}

// Material Table Relations
type MaterialRelation platformpostgres.RelationName

const (
	MaterialRelation_ParentSubShelf MaterialRelation = "ParentSubShelf"
)

/* ============================== Relative Type Conversion ============================== */

func (m *Material) ToPrivateMaterial() *gqlmodels.PrivateMaterial {
	return &gqlmodels.PrivateMaterial{
		ID:               m.Id,
		ParentSubShelfID: m.ParentSubShelfId,
		Name:             m.Name,
		Size:             m.Size,
		ContentType:      *m.ContentType.ToContractable(),
		ParseMediaType:   m.ParseMediaType,
		DeletedAt:        m.DeletedAt,
		UpdatedAt:        m.UpdatedAt,
		CreatedAt:        m.CreatedAt,
	}
}
