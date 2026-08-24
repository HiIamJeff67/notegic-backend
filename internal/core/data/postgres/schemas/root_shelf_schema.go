package schemas

import (
	"time"

	"github.com/google/uuid"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type RootShelf struct {
	Id             uuid.UUID  `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	OwnerId        uuid.UUID  `json:"ownerId" gorm:"column:owner_id; type:uuid; not null;"`              // Previous unique-name constraint: uniqueIndex:shelf_idx_name_owner_id,where:deleted_at IS NULL
	Name           string     `json:"name" gorm:"column:name; size:128; not null; default:'undefined';"` // Previous unique-name constraint: uniqueIndex:shelf_idx_name_owner_id,where:deleted_at IS NULL
	SubShelfCount  int64      `json:"subShelfCount" gorm:"column:sub_shelf_count; type:bigint; not null; default:0;"`
	ItemCount      int64      `json:"itemCount" gorm:"column:item_count; type:bigint; not null; default:0;"`
	LastAnalyzedAt time.Time  `json:"lastAnalyzedAt" gorm:"column:last_analyzed_at; type:timestamptz; not null; default:NOW();"`
	DeletedAt      *time.Time `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	SubShelves     []SubShelf       `json:"subShelves" gorm:"foreignKey:RootShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Items          []Item           `json:"items" gorm:"foreignKey:RootShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToShelves []UsersToShelves `json:"usersToShelves" gorm:"foreignKey:RootShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Shelf Table Name
func (RootShelf) TableName() string {
	return "RootShelfTable"
}

// Shelf Table Relations
type RootShelfRelation platformpostgres.RelationName

const (
	RootShelfRelation_SubShelves          RootShelfRelation = "SubShelves"
	RootShelfRelation_Items               RootShelfRelation = "Items"
	RootShelfRelation_UsersToShelves      RootShelfRelation = "UsersToShelves"
	RootShelfRelation_UsersToShelves_User RootShelfRelation = "UsersToShelves.User"
)

/* ============================== Relative Type Conversion ============================== */

func (rs *RootShelf) ToPrivateRootShelf(permission enums.AccessControlPermission) *gqlmodels.PrivateRootShelf {
	itemIds := make([]uuid.UUID, len(rs.Items))
	for index, item := range rs.Items {
		itemIds[index] = item.Id
	}

	return &gqlmodels.PrivateRootShelf{
		ID:             rs.Id,
		Name:           rs.Name,
		Permission:     *permission.ToContractable(),
		SubShelfCount:  rs.SubShelfCount,
		ItemCount:      rs.ItemCount,
		LastAnalyzedAt: rs.LastAnalyzedAt,
		DeletedAt:      rs.DeletedAt,
		UpdatedAt:      rs.UpdatedAt,
		CreatedAt:      rs.CreatedAt,
		Owner:          &gqlmodels.PublicUser{},
		Sharers:        make([]*gqlmodels.PublicUser, 0),
		ItemIds:        itemIds,
	}
}
