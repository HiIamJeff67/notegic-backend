package schemas

import (
	"time"

	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type RoutinesToTags struct {
	RoutineId uuid.UUID `json:"routineId" gorm:"column:routine_id; type:uuid; primaryKey; index:routines_to_tags_idx_user_id_routine_id,priority:2;"`
	TagId     uuid.UUID `json:"tagId" gorm:"column:tag_id; type:uuid; primaryKey;"`
	UserId    uuid.UUID `json:"userId" gorm:"column:user_id; type:uuid; not null; index:routines_to_tags_idx_user_id_routine_id,priority:1; index:routines_to_tags_idx_user_id_station_id,priority:1;"`
	StationId uuid.UUID `json:"stationId" gorm:"column:station_id; type:uuid; not null; index:routines_to_tags_idx_user_id_station_id,priority:2;"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	Routine        Routine         `json:"routine" gorm:"foreignKey:RoutineId,StationId; references:Id,StationId; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Tag            RoutineTag      `json:"tag" gorm:"foreignKey:TagId,UserId; references:Id,OwnerId; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToStation UsersToStations `json:"usersToStation" gorm:"foreignKey:UserId,StationId; references:UserId,StationId; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// RoutinesToTags Table Name
func (RoutinesToTags) TableName() string {
	return postgres.TableName_RoutinesToTagsTable.String()
}

// RoutinesToTags Table Relations
type RoutinesToTagsRelation postgres.RelationName

const (
	RoutinesToTagsRelation_Routine        RoutinesToTagsRelation = "Routine"
	RoutinesToTagsRelation_Tag            RoutinesToTagsRelation = "Tag"
	RoutinesToTagsRelation_UsersToStation RoutinesToTagsRelation = "UsersToStation"
)
