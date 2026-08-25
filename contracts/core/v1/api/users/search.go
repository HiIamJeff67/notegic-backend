package apicontract

import (
	"github.com/google/uuid"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
)

type SearchUsersRequestDto = cgqlmodels.SearchUserInput
type SearchUsersResponseDto = cgqlmodels.SearchUserConnection
type LoadThemeAuthorsRequestDto []uuid.UUID
type LoadThemeAuthorsResponseDto []*cgqlmodels.PublicUser
