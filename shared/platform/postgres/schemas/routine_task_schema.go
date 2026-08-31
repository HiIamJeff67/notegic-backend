package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type RoutineTask struct {
	Id                uuid.UUID                 `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	RoutineId         uuid.UUID                 `json:"routineId" gorm:"column:routine_id; type:uuid; not null;"`
	ActorUserId       uuid.UUID                 `json:"actorUserId" gorm:"column:actor_user_id; type:uuid; not null;"`
	Title             string                    `json:"title" gorm:"column:title; size:128; not null; default:'undefined';"`
	Purpose           cenums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose; type:\"RoutineTaskPurpose\"; not null; default:'CreateBlockPack';"`
	Payload           datatypes.JSON            `json:"payload" gorm:"column:payload; type:jsonb; not null; default:'{}'; check:routine_task_check_payload_size,octet_length(payload::text) <= 16777216;"`
	CostUnit          int64                     `json:"costUnit" gorm:"column:cost_unit; type:bigint; not null; default:0; check:routine_task_check_cost_unit_non_negative,cost_unit >= 0;"`
	Priority          int32                     `json:"priority" gorm:"column:priority; type:integer; not null; default:0; check:routine_task_check_priority_validation,priority >= 0 AND priority <= 100;"`
	MaxAttempts       int32                     `json:"maxAttempts" gorm:"column:max_attempts; type:integer; not null; default:1; check:routine_task_check_max_attempts_non_negative,max_attempts > 0;"`
	RecordScheduledAt time.Time                 `json:"-" gorm:"-"` // execution-only pattern interpolation context
	RecordId          uuid.UUID                 `json:"-" gorm:"-"` // execution-only pattern interpolation context
	UpdatedAt         time.Time                 `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt         time.Time                 `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	// Routine and User are owned and migrated by Core. DurableJob may use these
	// relations at runtime, but it must not let AutoMigrate create their tables.
	Routine       Routine             `json:"routine" gorm:"-:migration;foreignKey:RoutineId;references:Id;constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	ActorUser     UserView            `json:"actorUser" gorm:"-:migration;foreignKey:ActorUserId;references:Id;constraint:OnUpdate:CASCADE, OnDelete:RESTRICT;"`
	Records       []RoutineTaskRecord `json:"records" gorm:"foreignKey:RoutineTaskId;references:Id;constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	PreviousTasks []RoutineTask       `json:"previousTasks" gorm:"many2many:RoutineDependencyTable;foreignKey:Id;references:Id;joinForeignKey:RoutineTaskId;joinReferences:PreviousRoutineTaskId;"`
	NextTasks     []RoutineTask       `json:"nextTasks" gorm:"many2many:RoutineDependencyTable;foreignKey:Id;references:Id;joinForeignKey:PreviousRoutineTaskId;joinReferences:RoutineTaskId;"`
}

// RoutineTask Table Name
func (RoutineTask) TableName() string {
	return postgres.TableName_RoutineTaskTable.String()
}

// RoutineTask Table Relations
type RoutineTaskRelation postgres.RelationName

const (
	RoutineTaskRelation_Routine       RoutineTaskRelation = "Routine"
	RoutineTaskRelation_ActorUser     RoutineTaskRelation = "ActorUser"
	RoutineTaskRelation_Records       RoutineTaskRelation = "Records"
	RoutineTaskRelation_PreviousTasks RoutineTaskRelation = "PreviousTasks"
	RoutineTaskRelation_NextTasks     RoutineTaskRelation = "NextTasks"
)

/* ============================== Relative Type Conversion ============================== */

func (rt *RoutineTask) ToPrivateRoutineTask() *cgqlmodels.PrivateRoutineTask {
	previousTaskIds := make([]uuid.UUID, 0, len(rt.PreviousTasks))
	for _, previousTask := range rt.PreviousTasks {
		previousTaskIds = append(previousTaskIds, previousTask.Id)
	}
	return &cgqlmodels.PrivateRoutineTask{
		ID:                     rt.Id,
		RoutineID:              rt.RoutineId,
		Title:                  rt.Title,
		Purpose:                rt.Purpose,
		Payload:                json.RawMessage(rt.Payload),
		CostUnit:               rt.CostUnit,
		Priority:               rt.Priority,
		MaxAttempts:            rt.MaxAttempts,
		PreviousRoutineTaskIds: previousTaskIds,
		UpdatedAt:              rt.UpdatedAt,
		CreatedAt:              rt.CreatedAt,
	}
}
