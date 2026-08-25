package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UsersToShelves struct {
	UserId      uuid.UUID                     `json:"userId" gorm:"column:user_id; type:uuid; primaryKey;"`
	RootShelfId uuid.UUID                     `json:"rootShelfId" gorm:"column:root_shelf_id; type:uuid; primaryKey; uniqueIndex:idx_root_shelf_owner,where:permission = 'Owner';"`
	Permission  enums.AccessControlPermission `json:"permission" gorm:"column:permission; type:\"AccessControlPermission\"; not null; default:'Read';"`
	UpdatedAt   time.Time                     `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt   time.Time                     `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	User      User      `gorm:"foreignKey:UserId; reference:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RootShelf RootShelf `gorm:"foreignKey:RootShelfId; reference:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// UsersToShelves Table Name
func (UsersToShelves) TableName() string {
	return platformpostgres.TableName_UsersToShelvesTable.String()
}

// UsersToShelves Table Relations
type UsersToShelvesRelation platformpostgres.RelationName

const (
	UsersToShelvesRelation_User      UsersToShelvesRelation = "User"
	UsersToShelvesRelation_RootShelf UsersToShelvesRelation = "RootShelf"
)

/* ============================== Trigger Hooks ============================== */

func (uts *UsersToShelves) AfterSave(tx *gorm.DB) error {
	if uts.Permission == "Owner" {
		err := tx.Model(&RootShelf{Id: uts.RootShelfId}).
			Update("owner_id", uts.UserId).Error
		if err != nil {
			return err
		}
	}
	return nil
}
