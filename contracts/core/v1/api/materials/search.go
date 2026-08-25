package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchMaterialsRequestDto = cgqlmodels.SearchMaterialInput
type SearchMaterialsResponseDto = cgqlmodels.SearchMaterialConnection
