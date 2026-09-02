package inputs

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateRoutineTaskInput struct {
	ActorUserId uuid.UUID                 `json:"actorUserId" gorm:"column:actor_user_id;"`
	Title       string                    `json:"title" gorm:"column:title;"`
	Purpose     cenums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload     datatypes.JSON            `json:"payload" gorm:"column:payload;"`
	Priority    int32                     `json:"priority" gorm:"column:priority;"`
	MaxAttempts int32                     `json:"maxAttempts" gorm:"column:max_attempts;"`
}

type CreateRoutineTaskByRoutineIdInput struct {
	RoutineId   uuid.UUID                 `json:"routineId" gorm:"column:routine_id;"`
	ActorUserId uuid.UUID                 `json:"actorUserId" gorm:"column:actor_user_id;"`
	Title       string                    `json:"title" gorm:"column:title;"`
	Purpose     cenums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload     datatypes.JSON            `json:"payload" gorm:"column:payload;"`
	Priority    int32                     `json:"priority" gorm:"column:priority;"`
	MaxAttempts int32                     `json:"maxAttempts" gorm:"column:max_attempts;"`
}

type UpdateRoutineTaskInput struct {
	RoutineId   *uuid.UUID                 `json:"routineId" gorm:"column:routine_id;"`
	Title       *string                    `json:"title" gorm:"column:title;"`
	Purpose     *cenums.RoutineTaskPurpose `json:"purpose" gorm:"column:purpose;"`
	Payload     *datatypes.JSON            `json:"payload" gorm:"column:payload;"`
	Priority    *int32                     `json:"priority" gorm:"column:priority;"`
	MaxAttempts *int32                     `json:"maxAttempts" gorm:"column:max_attempts;"`
}

type PartialUpdateRoutineTaskInput = PartialUpdateInput[UpdateRoutineTaskInput]

type UpdateRoutineTaskByIdInput struct {
	Id                 uuid.UUID                                  `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineTaskInput] `json:"partialUpdateInput"`
}
