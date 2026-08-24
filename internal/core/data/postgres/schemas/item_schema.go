package schemas

import (
	"time"

	"github.com/google/uuid"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type Item struct {
	Id               uuid.UUID      `json:"id" gorm:"column:id; type:uuid; primaryKey;"`
	ParentSubShelfId uuid.UUID      `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id; type:uuid; not null;"`
	RootShelfId      uuid.UUID      `json:"rootShelfId" gorm:"column:root_shelf_id; type:uuid; not null;"`
	Type             enums.ItemType `json:"itemType" gorm:"column:type; type:\"ItemType\"; primaryKey;"`
	DeletedAt        *time.Time     `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt        time.Time      `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	ParentSubShelf  SubShelf          `json:"parentSubShelf" gorm:"foreignKey:ParentSubShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RootShelf       RootShelf         `json:"rootShelf" gorm:"foreignKey:RootShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutinesToItems []RoutinesToItems `json:"routinesToItems" gorm:"foreignKey:ItemId,ItemType; references:Id,Type; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Item Table Name
func (Item) TableName() string {
	return "ItemTable"
}

// Item Table Relations
type ItemRelation platformpostgres.RelationName

const (
	ItemRelation_ParentSubShelf  ItemRelation = "ParentSubShelf"
	ItemRelation_RootShelf       ItemRelation = "RootShelf"
	ItemRelation_RoutinesToItems ItemRelation = "RoutinesToItems"
)

/* ============================== Relative Type Conversion ============================== */

func (i *Item) ToPrivateItem() *gqlmodels.PrivateItem {
	routineIds := make([]uuid.UUID, 0, len(i.RoutinesToItems))
	for _, routineToItem := range i.RoutinesToItems {
		routineIds = append(routineIds, routineToItem.RoutineId)
	}

	return &gqlmodels.PrivateItem{
		ID:               i.Id,
		ParentSubShelfID: i.ParentSubShelfId,
		RootShelfID:      i.RootShelfId,
		Type:             *i.Type.ToContractable(),
		DeletedAt:        i.DeletedAt,
		UpdatedAt:        i.UpdatedAt,
		CreatedAt:        i.CreatedAt,
		RoutineIds:       routineIds,
	}
}
