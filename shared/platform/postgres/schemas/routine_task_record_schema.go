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

type RoutineTaskRecord struct {
	Id              uuid.UUID                          `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	RoutineRecordId uuid.UUID                          `json:"routineRecordId" gorm:"column:routine_record_id; type:uuid; not null; uniqueIndex:routine_task_record_idx_routine_record_id_task_id; index:routine_task_record_idx_routine_record_id_status,priority:1;"`
	RoutineTaskId   uuid.UUID                          `json:"routineTaskId" gorm:"column:routine_task_id; type:uuid; not null; uniqueIndex:routine_task_record_idx_routine_record_id_task_id; index:routine_task_record_idx_routine_task_id;"`
	Purpose         cenums.RoutineTaskPurpose          `json:"purpose" gorm:"column:purpose; type:\"RoutineTaskPurpose\"; not null; default:'CreateBlockPack';"`
	Status          cenums.RoutineTaskRecordStatus     `json:"status" gorm:"column:status; type:\"RoutineTaskRecordStatus\"; not null; default:'Waiting'; index:routine_task_record_idx_status; index:routine_task_record_idx_routine_record_id_status,priority:2;"`
	ErrorCode       *cenums.RoutineTaskRecordErrorCode `json:"errorCode" gorm:"column:error_code; type:\"RoutineTaskRecordErrorCode\"; default:null;"`
	ErrorReason     *string                            `json:"errorReason" gorm:"column:error_reason; type:varchar(256); default:null;"`
	CostUnit        int64                              `json:"costUnit" gorm:"column:cost_unit; type:bigint; not null; default:0; check:routine_task_record_cost_unit_non_negative,cost_unit >= 0;"`
	Attempts        int32                              `json:"attempts" gorm:"column:attempts; type:integer; not null; default:0; check:routine_task_record_attempts_non_negative,attempts >= 0;"`
	PayloadSnapshot datatypes.JSON                     `json:"payloadSnapshot" gorm:"column:payload_snapshot; type:jsonb; not null; default:'{}';"`
	ResultSnapshot  datatypes.JSON                     `json:"resultSnapshot" gorm:"column:result_snapshot; type:jsonb; not null; default:'{}';"`
	ActualStartedAt *time.Time                         `json:"actualStartedAt" gorm:"column:actual_started_at; type:timestamptz; default:null;"`
	ActualEndedAt   *time.Time                         `json:"actualEndedAt" gorm:"column:actual_ended_at; type:timestamptz; default:null;"`
	UpdatedAt       time.Time                          `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt       time.Time                          `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	RoutineRecord RoutineRecord `json:"routineRecord" gorm:"foreignKey:RoutineRecordId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Origin        RoutineTask   `json:"origin" gorm:"foreignKey:RoutineTaskId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (RoutineTaskRecord) TableName() string {
	return postgres.TableName_RoutineTaskRecordTable.String()
}

type RoutineTaskRecordRelation postgres.RelationName

const (
	RoutineTaskRecordRelation_RoutineRecord RoutineTaskRecordRelation = "RoutineRecord"
	RoutineTaskRecordRelation_Origin        RoutineTaskRecordRelation = "Origin"
)

func (rtr *RoutineTaskRecord) ToPrivateRoutineTaskRecord() *cgqlmodels.PrivateRoutineTaskRecord {
	return &cgqlmodels.PrivateRoutineTaskRecord{
		ID:              rtr.Id,
		RoutineRecordID: rtr.RoutineRecordId,
		RoutineTaskID:   rtr.RoutineTaskId,
		Purpose:         rtr.Purpose,
		Status:          rtr.Status,
		ErrorCode:       rtr.ErrorCode,
		ErrorReason:     rtr.ErrorReason,
		CostUnit:        rtr.CostUnit,
		Attempts:        rtr.Attempts,
		PayloadSnapshot: json.RawMessage(rtr.PayloadSnapshot),
		ResultSnapshot:  json.RawMessage(rtr.ResultSnapshot),
		ActualStartedAt: rtr.ActualStartedAt,
		ActualEndedAt:   rtr.ActualEndedAt,
		UpdatedAt:       rtr.UpdatedAt,
		CreatedAt:       rtr.CreatedAt,
	}
}
