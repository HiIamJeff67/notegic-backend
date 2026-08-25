package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/root-shelves"
)

type CreateRootShelfRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.CreatableRootShelf,
		struct{},
		struct{},
	]
}

type CreateRootShelfResponseDto struct {
	Id             uuid.UUID `json:"id"`
	LastAnalyzedAt time.Time `json:"lastAnalyzedAt"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateRootShelvesRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelves []coretypes.CreatableRootShelf `json:"insertedRootShelves" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type CreateRootShelvesResponseDto struct {
	Ids            []uuid.UUID `json:"ids"`
	LastAnalyzedAt time.Time   `json:"lastAnalyzedAt"`
	CreatedAt      time.Time   `json:"createdAt"`
}
