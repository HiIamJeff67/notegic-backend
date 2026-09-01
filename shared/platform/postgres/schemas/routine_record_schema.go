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

type RoutineRecord struct {
	Id                uuid.UUID                  `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	RoutineId         uuid.UUID                  `json:"routineId" gorm:"column:routine_id; type:uuid; not null; uniqueIndex:routine_record_idx_routine_id_scheduled_at,priority:1;"`
	DefinitionVersion int64                      `json:"-" gorm:"column:definition_version; type:bigint; not null; default:1; uniqueIndex:routine_record_idx_routine_id_scheduled_at,priority:3; check:routine_record_definition_version_positive,definition_version > 0;"`
	Status            cenums.RoutineRecordStatus `json:"status" gorm:"column:status; type:\"RoutineRecordStatus\"; not null; default:'Pending';"`
	ScheduledAt       time.Time                  `json:"scheduledAt" gorm:"column:scheduled_at; type:timestamptz; not null; uniqueIndex:routine_record_idx_routine_id_scheduled_at,priority:2;"`
	ActualStartedAt   *time.Time                 `json:"actualStartedAt" gorm:"column:actual_started_at; type:timestamptz; default:null;"`
	ActualEndedAt     *time.Time                 `json:"actualEndedAt" gorm:"column:actual_ended_at; type:timestamptz; default:null;"`
	TotalTaskCount    int32                      `json:"totalTaskCount" gorm:"column:total_task_count; type:integer; not null; default:0; check:routine_record_total_task_count_non_negative,total_task_count >= 0;"`
	SuccessTaskCount  int32                      `json:"successTaskCount" gorm:"column:success_task_count; type:integer; not null; default:0; check:routine_record_success_task_count_non_negative,success_task_count >= 0;"`
	FailedTaskCount   int32                      `json:"failedTaskCount" gorm:"column:failed_task_count; type:integer; not null; default:0; check:routine_record_failed_task_count_non_negative,failed_task_count >= 0;"`
	BlockedTaskCount  int32                      `json:"blockedTaskCount" gorm:"column:blocked_task_count; type:integer; not null; default:0; check:routine_record_blocked_task_count_non_negative,blocked_task_count >= 0;"`
	RunningTaskCount  int32                      `json:"runningTaskCount" gorm:"column:running_task_count; type:integer; not null; default:0; check:routine_record_running_task_count_non_negative,running_task_count >= 0;"`
	WaitingTaskCount  int32                      `json:"waitingTaskCount" gorm:"column:waiting_task_count; type:integer; not null; default:0; check:routine_record_waiting_task_count_non_negative,waiting_task_count >= 0;"`
	Snapshot          datatypes.JSON             `json:"snapshot" gorm:"column:snapshot; type:jsonb; not null; default:'{}';"`
	UpdatedAt         time.Time                  `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt         time.Time                  `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	Routine            Routine             `json:"routine" gorm:"-:migration;foreignKey:RoutineId; references:Id;"`
	RoutineTaskRecords []RoutineTaskRecord `json:"routineTaskRecords" gorm:"foreignKey:RoutineRecordId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (RoutineRecord) TableName() string {
	return postgres.TableName_RoutineRecordTable.String()
}

type RoutineRecordRelation postgres.RelationName

const (
	RoutineRecordRelation_Routine            RoutineRecordRelation = "Routine"
	RoutineRecordRelation_RoutineTaskRecords RoutineRecordRelation = "RoutineTaskRecords"
)

func (rr *RoutineRecord) ToPrivateRoutineRecord() *cgqlmodels.PrivateRoutineRecord {
	return &cgqlmodels.PrivateRoutineRecord{
		ID:               rr.Id,
		RoutineID:        rr.RoutineId,
		Status:           rr.Status,
		ScheduledAt:      rr.ScheduledAt,
		ActualStartedAt:  rr.ActualStartedAt,
		ActualEndedAt:    rr.ActualEndedAt,
		TotalTaskCount:   rr.TotalTaskCount,
		SuccessTaskCount: rr.SuccessTaskCount,
		FailedTaskCount:  rr.FailedTaskCount,
		BlockedTaskCount: rr.BlockedTaskCount,
		RunningTaskCount: rr.RunningTaskCount,
		WaitingTaskCount: rr.WaitingTaskCount,
		Snapshot:         json.RawMessage(rr.Snapshot),
		UpdatedAt:        rr.UpdatedAt,
		CreatedAt:        rr.CreatedAt,
	}
}
