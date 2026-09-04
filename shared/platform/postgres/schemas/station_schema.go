package schemas

import (
	"time"

	"github.com/google/uuid"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Station struct {
	Id                  uuid.UUID             `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	OwnerId             uuid.UUID             `json:"ownerId" gorm:"column:owner_id; type:uuid; not null;"`
	Name                string                `json:"name" gorm:"column:name; size:128; not null; default:'undefined';"` // Previous unique-name constraint: unique
	Description         string                `json:"description" gorm:"column:description; size:1024; not null; default:'';"`
	Icon                *cenums.SupportedIcon `json:"icon" gorm:"column:icon; type:\"SupportedIcon\"; default:null;"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" gorm:"column:header_background_url; default:null;"`
	RoutineCount        int64                 `json:"routineCount" gorm:"column:routine_count; type:bigint; not null; default:0; check:station_check_max_routine_count,routine_count <= 500;"`
	DeletedAt           *time.Time            `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt           time.Time             `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt           time.Time             `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	Owner           User              `json:"owner" gorm:"foreignKey:OwnerId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToStations []UsersToStations `json:"usersToStations" gorm:"foreignKey:StationId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Routines        []Routine         `json:"routines" gorm:"foreignKey:StationId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Station Table Name
func (Station) TableName() string {
	return postgres.TableName_StationTable.String()
}

// Station Table Relations
type StationRelation postgres.RelationName

const (
	StationRelation_Owner           StationRelation = "Owner"
	StationRelation_UsersToStations StationRelation = "UsersToStations"
	StationRelation_Routines        StationRelation = "Routines"
)

/* ============================== Relative Type Conversion ============================== */

func (s *Station) ToPrivateStation(permission cenums.AccessControlPermission) *cgqlmodels.PrivateStation {
	return &cgqlmodels.PrivateStation{
		ID:                  s.Id,
		Permission:          permission,
		Name:                s.Name,
		Description:         s.Description,
		Icon:                s.Icon,
		HeaderBackgroundURL: s.HeaderBackgroundURL,
		RoutineCount:        s.RoutineCount,
		DeletedAt:           s.DeletedAt,
		UpdatedAt:           s.UpdatedAt,
		CreatedAt:           s.CreatedAt,
	}
}

func (s *Station) ToPrivateSearchableStation(permission cenums.AccessControlPermission) *cgqlmodels.PrivateSearchableStation {
	return &cgqlmodels.PrivateSearchableStation{
		ID:                  s.Id,
		Permission:          permission,
		Name:                s.Name,
		Icon:                s.Icon,
		HeaderBackgroundURL: s.HeaderBackgroundURL,
		RoutineCount:        s.RoutineCount,
		DeletedAt:           s.DeletedAt,
		UpdatedAt:           s.UpdatedAt,
		CreatedAt:           s.CreatedAt,
	}
}
