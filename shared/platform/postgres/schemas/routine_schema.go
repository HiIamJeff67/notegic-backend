package schemas

import (
	"time"

	"github.com/google/uuid"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type Routine struct {
	Id                uuid.UUID             `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid(); uniqueIndex:routine_idx_id_station_id;"`
	StationId         uuid.UUID             `json:"stationId" gorm:"column:station_id; type:uuid; not null; uniqueIndex:routine_idx_id_station_id;"`
	Title             string                `json:"title" gorm:"column:title; size: 128; not null; default:'undefined';"`
	Description       string                `json:"description" gorm:"column:description; size:1024; not null; default:'';"`
	Status            cenums.RoutineStatus  `json:"status" gorm:"column:status; type:\"RoutineStatus\"; not null; default:'Scheduled';"`
	Phase             *cenums.RoutinePhase  `json:"phase" gorm:"column:phase; type:\"RoutinePhase\"; default:null;"`
	DefinitionVersion int64                 `json:"-" gorm:"column:definition_version; type:bigint; not null; default:1; check:routine_definition_version_positive,definition_version > 0;"`
	IsPinned          bool                  `json:"isPinned" gorm:"column:is_pinned; type:boolean; not null; default:false;"`
	ScheduledStartAt  time.Time             `json:"scheduledStartAt" gorm:"column:scheduled_start_at; type:timestamptz; not null; default:NOW();"`                 // check: routine_check_scheduled_start_minute_precision and routine_check_scheduled_time_in_period
	ScheduledEndAt    time.Time             `json:"scheduledEndAt" gorm:"column:scheduled_end_at; type:timestamptz; not null; default:NOW() + INTERVAL '1 hour';"` // check: routine_check_scheduled_end_minute_precision and routine_check_scheduled_time_in_period
	Period            *cenums.RoutinePeriod `json:"period" gorm:"column:period; type:\"RoutinePeriod\"; default:null;"`                                            // check: routine_check_scheduled_time_in_period
	Timezone          string                `json:"timezone" gorm:"column:timezone; size:64; not null; default:'UTC';"`                                            // validate by validation package with time.LoadLocation
	DeletedAt         *time.Time            `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt         time.Time             `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt         time.Time             `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	Station         Station           `json:"station" gorm:"foreignKey:StationId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutinesToTags  []RoutinesToTags  `json:"routinesToTags" gorm:"foreignKey:RoutineId,StationId; references:Id,StationId; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutineTasks    []RoutineTask     `json:"routineTasks" gorm:"foreignKey:RoutineId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutinesToItems []RoutinesToItems `json:"routinesToItems" gorm:"foreignKey:RoutineId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// Routine Table Name
func (Routine) TableName() string {
	return postgres.TableName_RoutineTable.String()
}

// Routine Table Relations
type RoutineRelation postgres.RelationName

const (
	RoutineRelation_Station         RoutineRelation = "Station"
	RoutineRelation_RoutinesToTags  RoutineRelation = "RoutinesToTags"
	RoutineRelation_RoutineTasks    RoutineRelation = "RoutineTasks"
	RoutineRelation_RoutinesToItems RoutineRelation = "RoutinesToItems"
)

/* ============================== Relative Type Conversion ============================== */

func (r *Routine) ToPrivateRoutine() *cgqlmodels.PrivateRoutine {
	tagIds := make([]uuid.UUID, 0, len(r.RoutinesToTags))
	for _, routineToTag := range r.RoutinesToTags {
		tagIds = append(tagIds, routineToTag.TagId)
	}

	taskIds := make([]uuid.UUID, 0, len(r.RoutineTasks))
	for _, routineTask := range r.RoutineTasks {
		taskIds = append(taskIds, routineTask.Id)
	}

	itemIds := make([]uuid.UUID, 0, len(r.RoutinesToItems))
	for _, routineToItem := range r.RoutinesToItems {
		itemIds = append(itemIds, routineToItem.ItemId)
	}

	return &cgqlmodels.PrivateRoutine{
		ID:               r.Id,
		StationID:        r.StationId,
		Title:            r.Title,
		Description:      r.Description,
		Status:           r.Status,
		Phase:            r.Phase,
		IsPinned:         r.IsPinned,
		ScheduledStartAt: r.ScheduledStartAt,
		ScheduledEndAt:   r.ScheduledEndAt,
		Period:           r.Period,
		Timezone:         r.Timezone,
		DeletedAt:        r.DeletedAt,
		UpdatedAt:        r.UpdatedAt,
		CreatedAt:        r.CreatedAt,
		TagIds:           tagIds,
		TaskIds:          taskIds,
		ItemIds:          itemIds,
	}
}

func (r *Routine) ToPrivateSearchableRoutine() *cgqlmodels.PrivateSearchableRoutine {
	tagIds := make([]uuid.UUID, 0, len(r.RoutinesToTags))
	for _, routineToTag := range r.RoutinesToTags {
		tagIds = append(tagIds, routineToTag.TagId)
	}

	taskIds := make([]uuid.UUID, 0, len(r.RoutineTasks))
	for _, routineTask := range r.RoutineTasks {
		taskIds = append(taskIds, routineTask.Id)
	}

	itemIds := make([]uuid.UUID, 0, len(r.RoutinesToItems))
	for _, routineToItem := range r.RoutinesToItems {
		itemIds = append(itemIds, routineToItem.ItemId)
	}

	return &cgqlmodels.PrivateSearchableRoutine{
		ID:               r.Id,
		StationID:        r.StationId,
		Title:            r.Title,
		Status:           r.Status,
		Phase:            r.Phase,
		IsPinned:         r.IsPinned,
		ScheduledStartAt: r.ScheduledStartAt,
		ScheduledEndAt:   r.ScheduledEndAt,
		Period:           r.Period,
		Timezone:         r.Timezone,
		DeletedAt:        r.DeletedAt,
		UpdatedAt:        r.UpdatedAt,
		CreatedAt:        r.CreatedAt,
		TagIds:           tagIds,
		TaskIds:          taskIds,
		ItemIds:          itemIds,
	}
}
