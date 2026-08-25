package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdateRoutineTaskRecordFailureInput struct {
	Id          uuid.UUID                         `json:"id" gorm:"column:id;"`
	ErrorCode   cenums.RoutineTaskRecordErrorCode `json:"errorCode" gorm:"column:error_code;"`
	ErrorReason string                            `json:"errorReason" gorm:"column:error_reason;"`
}
