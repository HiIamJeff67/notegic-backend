package inputs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type CreateRoutineTaskInput struct {
	ActorUserId     uuid.UUID                `json:"actorUserId" gorm:"column:actor_user_id;"`
	Title           string                   `json:"title" gorm:"column:title;"`
	Purpose         enums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload         datatypes.JSON           `json:"payload" gorm:"column:payload;"`
	Priority        int32                    `json:"priority" gorm:"column:priority;"`
	MaxAttempts     int32                    `json:"maxAttempts" gorm:"column:max_attempts;"`
	Period          *enums.RoutinePeriod     `json:"period" gorm:"column:period;"`
	NextScheduledAt time.Time                `json:"nextScheduledAt" gorm:"column:next_scheduled_at;"`
	ScheduledAt     time.Time                `json:"scheduledAt" gorm:"column:scheduled_at;"`
}

type CreateRoutineTaskByRoutineIdInput struct {
	RoutineId       uuid.UUID                `json:"routineId" gorm:"column:routine_id;"`
	ActorUserId     uuid.UUID                `json:"actorUserId" gorm:"column:actor_user_id;"`
	Title           string                   `json:"title" gorm:"column:title;"`
	Purpose         enums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload         datatypes.JSON           `json:"payload" gorm:"column:payload;"`
	Priority        int32                    `json:"priority" gorm:"column:priority;"`
	MaxAttempts     int32                    `json:"maxAttempts" gorm:"column:max_attempts;"`
	Period          *enums.RoutinePeriod     `json:"period" gorm:"column:period;"`
	NextScheduledAt time.Time                `json:"nextScheduledAt" gorm:"column:next_scheduled_at;"`
	ScheduledAt     time.Time                `json:"scheduledAt" gorm:"column:scheduled_at;"`
}

type UpdateRoutineTaskInput struct {
	RoutineId       *uuid.UUID                `json:"routineId" gorm:"column:routine_id;"`
	Title           *string                   `json:"title" gorm:"column:title;"`
	Purpose         *enums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload         *datatypes.JSON           `json:"payload" gorm:"column:payload;"`
	Priority        *int32                    `json:"priority" gorm:"column:priority;"`
	MaxAttempts     *int32                    `json:"maxAttempts" gorm:"column:max_attempts;"`
	Period          *enums.RoutinePeriod      `json:"period" gorm:"column:period;"`
	NextScheduledAt *time.Time                `json:"nextScheduledAt" gorm:"column:next_scheduled_at;"`
	ScheduledAt     *time.Time                `json:"scheduledAt" gorm:"column:scheduled_at;"`
}

type PartialUpdateRoutineTaskInput = PartialUpdateInput[UpdateRoutineTaskInput]

type UpdateRoutineTaskByIdInput struct {
	Id                 uuid.UUID                                  `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineTaskInput] `json:"partialUpdateInput"`
}
