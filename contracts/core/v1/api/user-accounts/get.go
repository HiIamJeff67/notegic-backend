package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyAccountRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type GetMyAccountResponseDto struct {
	CountryCode              *string   `json:"countryCode"`
	PhoneNumber              *string   `json:"phoneNumber"`
	GoogleCredential         *string   `json:"googleCrendential"`
	DiscordCredential        *string   `json:"discordCrendential"`
	RootShelfCount           int64     `json:"rootShelfCount"`
	BlockPackCount           int64     `json:"blockPackCount"`
	BlockCount               int64     `json:"blockCount"`
	MaterialCount            int64     `json:"materialCount"`
	WorkflowCount            int64     `json:"workflowCount"`
	AdditionalItemCount      int64     `json:"additionalItemCount"`
	StationCount             int64     `json:"stationCount"`
	RoutineCount             int64     `json:"routineCount"`
	RoutineTaskCostUnitCount int64     `json:"routineTaskCostUnitCount"`
	RoutineTagCount          int64     `json:"routineTagCount"`
	UpdatedAt                time.Time `json:"updatedAt"`
}
