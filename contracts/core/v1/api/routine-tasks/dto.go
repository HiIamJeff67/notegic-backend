package apicontract

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskResponseDto struct {
	Id              uuid.UUID                `json:"id"`
	RoutineId       uuid.UUID                `json:"routineId"`
	Title           string                   `json:"title"`
	Purpose         enums.RoutineTaskPurpose `json:"purpose"`
	Payload         datatypes.JSON           `json:"payload"`
	CostUnit        int64                    `json:"costUnit"`
	Priority        int32                    `json:"priority"`
	Status          enums.RoutineTaskStatus  `json:"status"`
	Attempts        int32                    `json:"attempts"`
	MaxAttempts     int32                    `json:"maxAttempts"`
	Period          *enums.RoutinePeriod     `json:"period"`
	NextScheduledAt time.Time                `json:"nextScheduledAt"`
	ScheduledAt     time.Time                `json:"scheduledAt"`
	ActualStartedAt *time.Time               `json:"actualStartedAt"`
	ActualEndedAt   *time.Time               `json:"actualEndedAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
	CreatedAt       time.Time                `json:"createdAt"`
}
