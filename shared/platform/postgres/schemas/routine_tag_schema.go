package schemas

import (
	"time"

	"github.com/google/uuid"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type RoutineTag struct {
	Id        uuid.UUID             `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid(); uniqueIndex:routine_tag_idx_id_owner_id,priority:1;"`
	OwnerId   uuid.UUID             `json:"ownerId" gorm:"column:owner_id; type:uuid; not null; index; uniqueIndex:routine_tag_idx_id_owner_id,priority:2;"`
	Name      string                `json:"name" gorm:"column:name; size: 128; not null; default:'undefined';"`
	Color     string                `json:"color" gorm:"column:color; size:7; not null; default:'#FFFFFF'; check:routine_tag_check_color_hex_code,color ~ '^#[0-9A-Fa-f]{6}$';"`
	Icon      *cenums.SupportedIcon `json:"icon" gorm:"column:icon; type:\"SupportedIcon\"; default:null;"`
	UpdatedAt time.Time             `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt time.Time             `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	Owner          User             `json:"owner" gorm:"foreignKey:OwnerId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutinesToTags []RoutinesToTags `json:"routinesToTags" gorm:"foreignKey:TagId,UserId; references:Id,OwnerId; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// RoutineTag Table Name
func (RoutineTag) TableName() string {
	return postgres.TableName_RoutineTagTable.String()
}

// RoutineTag Table Relations
type RoutineTagRelation postgres.RelationName

const (
	RoutineTagRelation_Owner          RoutineTagRelation = "Owner"
	RoutineTagRelation_RoutinesToTags RoutineTagRelation = "RoutinesToTags"
)

/* ============================== Relative Type Conversion ============================== */

func (rt *RoutineTag) ToPrivateRoutineTag() *cgqlmodels.PrivateRoutineTag {
	return &cgqlmodels.PrivateRoutineTag{
		ID:        rt.Id,
		Name:      rt.Name,
		Color:     rt.Color,
		Icon:      rt.Icon,
		UpdatedAt: rt.UpdatedAt,
		CreatedAt: rt.CreatedAt,
	}
}
