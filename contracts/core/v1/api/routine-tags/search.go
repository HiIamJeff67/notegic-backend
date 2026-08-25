package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchRoutineTagsRequestDto = cgqlmodels.SearchRoutineTagInput
type SearchRoutineTagsResponseDto = cgqlmodels.SearchRoutineTagConnection
