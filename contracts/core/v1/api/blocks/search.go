package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchBlocksRequestDto = cgqlmodels.SearchBlockInput
type SearchBlocksResponseDto = cgqlmodels.SearchBlockConnection
