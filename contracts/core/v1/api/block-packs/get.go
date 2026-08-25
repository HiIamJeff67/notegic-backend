package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type GetMyBlockPackByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
			IsDeleted   *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type BlockPackResponseDto struct {
	Id                     uuid.UUID             `json:"id"`
	ParentSubShelfId       uuid.UUID             `json:"parentSubShelfId"`
	Name                   string                `json:"name"`
	Icon                   *cenums.SupportedIcon `json:"icon"`
	HeaderBackgroundURL    *string               `json:"headerBackgroundURL"`
	BlockCount             int64                 `json:"blockCount"`
	LastUpdateSequence     int64                 `json:"lastUpdateSequence"`
	CompactedUntilSequence int64                 `json:"compactedUntilSequence"`
	ProjectedUntilSequence int64                 `json:"projectedUntilSequence"`
	IsProjectionCurrent    bool                  `json:"isProjectionCurrent"`
	DeletedAt              *time.Time            `json:"deletedAt"`
	UpdatedAt              time.Time             `json:"updatedAt"`
	CreatedAt              time.Time             `json:"createdAt"`
}

type GetMyBlockPackByIdResponseDto = BlockPackResponseDto

type GetMyBlockPackAndItsParentByIdRequestDto = GetMyBlockPackByIdRequestDto

type BlockPackAndParentResponseDto struct {
	Id                           uuid.UUID                      `json:"id"`
	Name                         string                         `json:"name"`
	Icon                         *cenums.SupportedIcon          `json:"icon"`
	HeaderBackgroundURL          *string                        `json:"headerBackgroundURL"`
	BlockCount                   int64                          `json:"blockCount"`
	LastUpdateSequence           int64                          `json:"lastUpdateSequence"`
	CompactedUntilSequence       int64                          `json:"compactedUntilSequence"`
	ProjectedUntilSequence       int64                          `json:"projectedUntilSequence"`
	IsProjectionCurrent          bool                           `json:"isProjectionCurrent"`
	DeletedAt                    *time.Time                     `json:"deletedAt"`
	UpdatedAt                    time.Time                      `json:"updatedAt"`
	CreatedAt                    time.Time                      `json:"createdAt"`
	RootShelfId                  uuid.UUID                      `json:"rootShelfId"`
	Permission                   cenums.AccessControlPermission `json:"permission"`
	ParentSubShelfId             uuid.UUID                      `json:"parentSubShelfId"`
	ParentSubShelfPrevSubShelfId *uuid.UUID                     `json:"parentSubShelfPrevSubShelfId"`
	ParentSubShelfName           string                         `json:"parentSubShelfName"`
	ParentSubShelfPath           []uuid.UUID                    `json:"parentSubShelfPath"`
	ParentSubShelfDeletedAt      *time.Time                     `json:"parentSubShelfDeletedAt"`
	ParentSubShelfUpdatedAt      time.Time                      `json:"parentSubShelfUpdatedAt"`
	ParentSubShelfCreatedAt      time.Time                      `json:"parentSubShelfCreatedAt"`
}

type GetMyBlockPackAndItsParentByIdResponseDto = BlockPackAndParentResponseDto

type GetMyBlockPacksByParentSubShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			ParentSubShelfId uuid.UUID `json:"parentSubShelfId" validate:"required"`
			AreDeleted       *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMyBlockPacksByParentSubShelfIdResponseDto []BlockPackResponseDto

type GetAllMyBlockPacksByRootShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
			AreDeleted  *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetAllMyBlockPacksByRootShelfIdResponseDto []BlockPackResponseDto
