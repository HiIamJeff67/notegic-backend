package apicontract

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyBlockByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			BlockId uuid.UUID `json:"blockId" validate:"required"`
		},
		struct{},
	]
}

type BlockResponseDto struct {
	Id            uuid.UUID       `json:"id"`
	BlockPackId   uuid.UUID       `json:"blockPackId"`
	ParentBlockId *uuid.UUID      `json:"parentBlockId"`
	PrevBlockId   *uuid.UUID      `json:"prevBlockId"`
	NextBlockId   *uuid.UUID      `json:"nextBlockId"`
	Type          string          `json:"type"`
	Props         json.RawMessage `json:"props"`
	Content       json.RawMessage `json:"content"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type GetMyBlockByIdResponseDto = BlockResponseDto

type GetMyBlocksByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			BlockIds []uuid.UUID `json:"blockIds" form:"blockIds" validate:"required,min=1,max=1024"`
		},
		struct{},
	]
}

type GetMyBlocksByIdsResponseDto []BlockResponseDto

type GetMyBlocksByBlockPackIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
		},
		struct{},
	]
}

type GetMyBlocksByBlockPackIdResponseDto []BlockResponseDto
