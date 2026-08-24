package schemas

import (
	"time"

	"github.com/google/uuid"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type RoutinesToItems struct {
	RoutineId uuid.UUID      `json:"routineId" gorm:"column:routine_id; type:uuid; primaryKey;"`
	ItemId    uuid.UUID      `json:"itemId" gorm:"column:item_id; type:uuid; primaryKey;"`
	ItemType  enums.ItemType `json:"itemType" gorm:"column:type; type:\"ItemType\"; primaryKey;"`
	CreatedAt time.Time      `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	Routine Routine `json:"routine" gorm:"foreignKey:RoutineId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Item    Item    `json:"item" gorm:"foreignKey:ItemId,ItemType; references:Id,Type; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// RoutinesToItems Table Name
func (RoutinesToItems) TableName() string {
	return "RoutinesToItemsTable"
}

// RoutinesToItems Table Relations
type RoutinesToItemsRelation platformpostgres.RelationName

const (
	RoutinesToItemsRelation_Routine RoutinesToItemsRelation = "Routine"
	RoutinesToItemsRelation_Item    RoutinesToItemsRelation = "Item"
)
