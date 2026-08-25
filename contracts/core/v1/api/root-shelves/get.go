package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyRootShelfByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
			IsDeleted   *bool     `json:"isDeleted,omitempty" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMyRootShelfByIdResponseDto struct {
	Id             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Permission     string     `json:"permission"`
	SubShelfCount  int64      `json:"subShelfCount"`
	ItemCount      int64      `json:"itemCount"`
	LastAnalyzedAt time.Time  `json:"lastAnalyzedAt"`
	DeletedAt      *time.Time `json:"deletedAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}
