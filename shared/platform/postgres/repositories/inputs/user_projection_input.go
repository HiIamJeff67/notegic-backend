package inputs

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateUserProjectionInput struct {
	PublicId uuid.UUID
	Plan     cenums.UserPlan
	Status   cenums.UserStatus
}
