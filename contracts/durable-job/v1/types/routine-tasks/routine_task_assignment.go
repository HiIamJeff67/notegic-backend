package routinetasktypes

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineTaskAssignment struct {
	RoutineTaskId       uuid.UUID                `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID                `json:"routineTaskRecordId"`
	RoutineId           uuid.UUID                `json:"routineId"`
	ActorUserId         uuid.UUID                `json:"actorUserId"`
	ActorUserPublicId   uuid.UUID                `json:"actorUserPublicId"`
	Title               string                   `json:"title"`
	Purpose             enums.RoutineTaskPurpose `json:"purpose"`
	Payload             json.RawMessage          `json:"payload"`
	CostUnit            int64                    `json:"costUnit"`
	Priority            int32                    `json:"priority"`
	Attempt             int32                    `json:"attempt"`
	ScheduledAt         time.Time                `json:"scheduledAt"`
	StartedAt           time.Time                `json:"startedAt"`
	PatternValues       map[string]string        `json:"patternValues,omitempty"`
}
