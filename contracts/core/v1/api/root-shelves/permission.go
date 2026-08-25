package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RootShelfPermissionResponseDto struct {
	UserPublicId uuid.UUID `json:"userPublicId"`
	Permission   string    `json:"permission"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GetMyRootShelfPermissionRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RootShelfId  uuid.UUID `json:"rootShelfId" validate:"required"`
			UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
		},
		struct{},
	]
}

type GetMyRootShelfPermissionResponseDto = RootShelfPermissionResponseDto

type CreateMyRootShelfPermissionRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Permission string `json:"permission" validate:"required,isaccesscontrolpermission"`
		},
		struct {
			RootShelfId  uuid.UUID `json:"rootShelfId" validate:"required"`
			UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
		},
		struct{},
	]
}

type CreateMyRootShelfPermissionResponseDto = RootShelfPermissionResponseDto
type UpsertMyRootShelfPermissionRequestDto = CreateMyRootShelfPermissionRequestDto
type UpsertMyRootShelfPermissionResponseDto = RootShelfPermissionResponseDto
type UpdateMyRootShelfPermissionRequestDto = CreateMyRootShelfPermissionRequestDto
type UpdateMyRootShelfPermissionResponseDto = RootShelfPermissionResponseDto

type UpsertableRootShelfPermission struct {
	UserPublicId uuid.UUID `json:"userPublicId" validate:"required"`
	Permission   string    `json:"permission" validate:"required,isaccesscontrolpermission"`
}

type UpsertMyRootShelfPermissionsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Permissions []UpsertableRootShelfPermission `json:"permissions" validate:"required,min=1,max=1024,dive"`
		},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
	]
}

type UpsertMyRootShelfPermissionsResponseDto struct {
	Permissions []RootShelfPermissionResponseDto `json:"permissions"`
}

type DeleteMyRootShelfPermissionRequestDto = GetMyRootShelfPermissionRequestDto
type DeleteMyRootShelfPermissionResponseDto struct{}

type DeleteMyRootShelfPermissionsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UserPublicIds []uuid.UUID `json:"userPublicIds" validate:"required,min=1,max=1024,dive,required"`
		},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
	]
}

type DeleteMyRootShelfPermissionsResponseDto struct{}
