package inputs

import (
	"time"

	"github.com/google/uuid"
)

type ConsumeRoutineTaskCostUnitInput struct {
	RoutineTaskId uuid.UUID `json:"routineTaskId"`
	UserId        uuid.UUID `json:"userId"`
	CostUnit      int64     `json:"costUnit"`
	Priority      int32     `json:"priority"`
	ScheduledAt   time.Time `json:"scheduledAt"`
}
