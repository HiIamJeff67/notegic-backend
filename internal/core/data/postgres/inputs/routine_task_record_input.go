package inputs

import (
	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type BulkCheckRoutineTaskRecordPermissionInput struct {
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
}

type UpdateRoutineTaskRecordFailureInput struct {
	Id          uuid.UUID                        `json:"id" gorm:"column:id;"`
	ErrorCode   enums.RoutineTaskRecordErrorCode `json:"errorCode" gorm:"column:error_code;"`
	ErrorReason string                           `json:"errorReason" gorm:"column:error_reason;"`
}
