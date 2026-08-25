package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type TransferMyRootShelfOwnershipRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			TargetUserPublicId uuid.UUID `json:"targetUserPublicId" validate:"required"`
		},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
	]
}

type TransferMyRootShelfOwnershipResponseDto struct {
	RootShelfId               uuid.UUID `json:"rootShelfId"`
	PreviousOwnerUserPublicId uuid.UUID `json:"previousOwnerUserPublicId"`
	NewOwnerUserPublicId      uuid.UUID `json:"newOwnerUserPublicId"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}
