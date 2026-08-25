package apicontract

import (
	"github.com/google/uuid"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
)

type LoadUserBadgesRequestDto []uuid.UUID
type LoadUserBadgesResponseDto []*cgqlmodels.PublicBadge
