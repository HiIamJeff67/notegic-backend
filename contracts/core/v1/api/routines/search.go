package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchRoutinesRequestDto = cgqlmodels.SearchRoutineInput
type SearchRoutinesResponseDto = cgqlmodels.SearchRoutineConnection
