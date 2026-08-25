package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RestoreMyRootShelfByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type RestoreMyRootShelfByIdResponseDto struct {
	Id             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	SubShelfCount  int64      `json:"subShelfCount"`
	ItemCount      int64      `json:"itemCount"`
	LastAnalyzedAt time.Time  `json:"lastAnalyzedAt"`
	DeletedAt      *time.Time `json:"deletedAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type RestoreMyRootShelvesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelfIds []uuid.UUID `json:"rootShelfIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type RestoreMyRootShelvesByIdsResponseDto []RestoreMyRootShelfByIdResponseDto
