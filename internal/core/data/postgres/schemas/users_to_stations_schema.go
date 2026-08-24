package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type UsersToStations struct {
	UserId     uuid.UUID                     `json:"userId" gorm:"column:user_id; type:uuid; primaryKey;"`
	StationId  uuid.UUID                     `json:"stationId" gorm:"column:station_id; type:uuid; primaryKey; uniqueIndex:idx_station_owner,where:permission = 'Owner';"`
	Permission enums.AccessControlPermission `json:"permission" gorm:"column:permission; type:\"AccessControlPermission\"; not null; default:'Read';"`
	UpdatedAt  time.Time                     `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt  time.Time                     `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	User    User    `json:"user" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Station Station `json:"station" gorm:"foreignKey:StationId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// UsersToStations Table Name
func (UsersToStations) TableName() string {
	return "UsersToStationsTable"
}

// UsersToStations Table Relations
type UsersToStationsRelation platformpostgres.RelationName

const (
	UsersToStationsRelation_User    UsersToStationsRelation = "User"
	UsersToStationsRelation_Station UsersToStationsRelation = "Station"
)

/* ============================== Trigger Hooks ============================== */

func (uts *UsersToStations) AfterSave(tx *gorm.DB) error {
	if uts.Permission != enums.AccessControlPermission_Owner {
		return nil
	}

	return tx.
		Model(&Station{Id: uts.StationId}).
		Update("owner_id", uts.UserId).Error
}
