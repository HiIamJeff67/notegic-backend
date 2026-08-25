package apicontract

import cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

type SearchThemesRequestDto = cgqlmodels.SearchThemeInput
type SearchThemesResponseDto = cgqlmodels.SearchThemeConnection
