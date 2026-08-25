package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchItemsRequestDto = cgqlmodels.SearchItemInput
type SearchItemsResponseDto = cgqlmodels.SearchItemConnection
