package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type MaterialResponseDto struct {
	Id               uuid.UUID                  `json:"id"`
	ParentSubShelfId uuid.UUID                  `json:"parentSubShelfId"`
	Name             string                     `json:"name"`
	Size             int64                      `json:"size"`
	ContentType      cenums.MaterialContentType `json:"contentType"`
	ParseMediaType   string                     `json:"parseMediaType"`
	DownloadURL      string                     `json:"downloadURL"`
	DeletedAt        *time.Time                 `json:"deletedAt"`
	UpdatedAt        time.Time                  `json:"updatedAt"`
	CreatedAt        time.Time                  `json:"createdAt"`
}

type GetMyMaterialByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			MaterialId uuid.UUID `json:"materialId" validate:"required"`
			IsDeleted  *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMyMaterialByIdResponseDto = MaterialResponseDto
type GetMyMaterialAndItsParentByIdRequestDto = GetMyMaterialByIdRequestDto

type MaterialAndParentResponseDto struct {
	Id                           uuid.UUID                  `json:"id"`
	Name                         string                     `json:"name"`
	Size                         int64                      `json:"size"`
	ContentType                  cenums.MaterialContentType `json:"contentType"`
	ParseMediaType               string                     `json:"parseMediaType"`
	DownloadURL                  string                     `json:"downloadURL"`
	DeletedAt                    *time.Time                 `json:"deletedAt"`
	UpdatedAt                    time.Time                  `json:"updatedAt"`
	CreatedAt                    time.Time                  `json:"createdAt"`
	RootShelfId                  uuid.UUID                  `json:"rootShelfId"`
	ParentSubShelfId             uuid.UUID                  `json:"parentSubShelfId"`
	ParentSubShelfPrevSubShelfId *uuid.UUID                 `json:"parentSubShelfPrevSubShelfId"`
	ParentSubShelfName           string                     `json:"parentSubShelfName"`
	ParentSubShelfPath           []uuid.UUID                `json:"parentSubShelfPath"`
	ParentSubShelfDeletedAt      *time.Time                 `json:"parentSubShelfDeletedAt"`
	ParentSubShelfUpdatedAt      time.Time                  `json:"parentSubShelfUpdatedAt"`
	ParentSubShelfCreatedAt      time.Time                  `json:"parentSubShelfCreatedAt"`
}

type GetMyMaterialAndItsParentByIdResponseDto = MaterialAndParentResponseDto

type GetMyMaterialsByParentSubShelfIdRequestDto struct {
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

type GetMyMaterialsByParentSubShelfIdResponseDto []MaterialResponseDto

type GetAllMyMaterialsByRootShelfIdRequestDto struct {
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

type GetAllMyMaterialsByRootShelfIdResponseDto []MaterialResponseDto
