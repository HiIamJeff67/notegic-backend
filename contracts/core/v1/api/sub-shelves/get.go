package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type SubShelfResponseDto struct {
	Id             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	RootShelfId    uuid.UUID   `json:"rootShelfId"`
	PrevSubShelfId *uuid.UUID  `json:"prevSubShelfId"`
	Path           []uuid.UUID `json:"path"`
	DeletedAt      *time.Time  `json:"deletedAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
	CreatedAt      time.Time   `json:"createdAt"`
}

type GetMySubShelfByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			SubShelfId uuid.UUID `json:"subShelfId" validate:"required"`
			IsDeleted  *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMySubShelfByIdResponseDto = SubShelfResponseDto

type GetMySubShelvesByPrevSubShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			PrevSubShelfId uuid.UUID `json:"prevSubShelfId" validate:"required"`
			AreDeleted     *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMySubShelvesByPrevSubShelfIdResponseDto []SubShelfResponseDto

type GetMySubShelvesByRootShelfIdRequestDto struct {
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

type GetMySubShelvesByRootShelfIdResponseDto []SubShelfResponseDto

type GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			PrevSubShelfId uuid.UUID `json:"prevSubShelfId" validate:"required"`
			AreDeleted     *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type SubShelfMaterialResponseDto struct {
	Id               uuid.UUID  `json:"id"`
	ParentSubShelfId uuid.UUID  `json:"parentSubShelfId"`
	Name             string     `json:"name"`
	Size             int64      `json:"size"`
	ContentType      string     `json:"contentType"`
	ParseMediaType   string     `json:"parseMediaType"`
	DownloadUrl      string     `json:"downloadURL"`
	DeletedAt        *time.Time `json:"deletedAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type SubShelfBlockPackResponseDto struct {
	Id                     uuid.UUID  `json:"id"`
	ParentSubShelfId       uuid.UUID  `json:"parentSubShelfId"`
	Name                   string     `json:"name"`
	Icon                   *string    `json:"icon"`
	HeaderBackgroundUrl    *string    `json:"headerBackgroundURL"`
	BlockCount             int64      `json:"blockCount"`
	LastUpdateSequence     int64      `json:"lastUpdateSequence"`
	CompactedUntilSequence int64      `json:"compactedUntilSequence"`
	ProjectedUntilSequence int64      `json:"projectedUntilSequence"`
	IsProjectionCurrent    bool       `json:"isProjectionCurrent"`
	DeletedAt              *time.Time `json:"deletedAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
	CreatedAt              time.Time  `json:"createdAt"`
}

type GetMySubShelvesAndItemsByPrevSubShelfIdResponseDto struct {
	SubShelves []SubShelfResponseDto          `json:"subShelves"`
	Materials  []SubShelfMaterialResponseDto  `json:"materials"`
	BlockPacks []SubShelfBlockPackResponseDto `json:"blockPacks"`
}
