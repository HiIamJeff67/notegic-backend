package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type StationPermissionResponseDto struct {
	UserPublicId uuid.UUID `json:"userPublicId"`
	Permission   string    `json:"permission"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GetMyStationPermissionRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			StationId    uuid.UUID `json:"stationId" validate:"required"`
			UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
		},
		struct{},
	]
}

type GetMyStationPermissionResponseDto = StationPermissionResponseDto

type CreateMyStationPermissionRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Permission string `json:"permission" validate:"required,isaccesscontrolpermission"`
		},
		struct {
			StationId    uuid.UUID `json:"stationId" validate:"required"`
			UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
		},
		struct{},
	]
}

type CreateMyStationPermissionResponseDto = StationPermissionResponseDto
type UpsertMyStationPermissionRequestDto = CreateMyStationPermissionRequestDto
type UpsertMyStationPermissionResponseDto = StationPermissionResponseDto
type UpdateMyStationPermissionRequestDto = CreateMyStationPermissionRequestDto
type UpdateMyStationPermissionResponseDto = StationPermissionResponseDto

type UpsertableStationPermission struct {
	UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
	Permission   string    `json:"permission" validate:"required,isaccesscontrolpermission"`
}

type UpsertMyStationPermissionsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Permissions []UpsertableStationPermission `json:"permissions" validate:"required,min=1,max=1024,dive"`
		},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
	]
}

type UpsertMyStationPermissionsResponseDto struct {
	Permissions []StationPermissionResponseDto `json:"permissions"`
}

type DeleteMyStationPermissionRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			StationId    uuid.UUID `json:"stationId" validate:"required"`
			UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
		},
		struct{},
	]
}

type DeleteMyStationPermissionResponseDto struct{}

type DeleteMyStationPermissionsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UserPublicIds []uuid.UUID `json:"userPublicIds" validate:"required,min=1,max=1024,dive,required"`
		},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
	]
}

type DeleteMyStationPermissionsResponseDto struct{}
