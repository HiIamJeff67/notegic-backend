package apicontract

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskResponseDto struct {
	Id                     uuid.UUID                 `json:"id"`
	RoutineId              uuid.UUID                 `json:"routineId"`
	Title                  string                    `json:"title"`
	Purpose                cenums.RoutineTaskPurpose `json:"purpose"`
	Payload                datatypes.JSON            `json:"payload"`
	CostUnit               int64                     `json:"costUnit"`
	Priority               int32                     `json:"priority"`
	MaxAttempts            int32                     `json:"maxAttempts"`
	PreviousRoutineTaskIds []uuid.UUID               `json:"previousRoutineTaskIds"`
	UpdatedAt              time.Time                 `json:"updatedAt"`
	CreatedAt              time.Time                 `json:"createdAt"`
}
