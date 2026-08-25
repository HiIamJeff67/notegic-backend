package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type BlockPackControllerInterface interface {
	GetMyBlockPackById(ctx *gin.Context, requestDto *capi.GetMyBlockPackByIdRequestDto)
	GetMyBlockPackAndItsParentById(ctx *gin.Context, requestDto *capi.GetMyBlockPackAndItsParentByIdRequestDto)
	GetMyBlockPacksByParentSubShelfId(ctx *gin.Context, requestDto *capi.GetMyBlockPacksByParentSubShelfIdRequestDto)
	GetAllMyBlockPacksByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMyBlockPacksByRootShelfIdRequestDto)
	CreateBlockPack(ctx *gin.Context, requestDto *capi.CreateBlockPackRequestDto)
	CreateBlockPacks(ctx *gin.Context, requestDto *capi.CreateBlockPacksRequestDto)
	UpdateMyBlockPackById(ctx *gin.Context, requestDto *capi.UpdateMyBlockPackByIdRequestDto)
	UpdateMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.UpdateMyBlockPacksByIdsRequestDto)
	MoveMyBlockPackByParentSubShelfId(ctx *gin.Context, requestDto *capi.MoveMyBlockPackByParentSubShelfIdRequestDto)
	MoveMyBlockPacksByParentSubShelfId(ctx *gin.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdRequestDto)
	MoveMyBlockPacksByParentSubShelfIds(ctx *gin.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto)
	RestoreMyBlockPackById(ctx *gin.Context, requestDto *capi.RestoreMyBlockPackByIdRequestDto)
	RestoreMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.RestoreMyBlockPacksByIdsRequestDto)
	DeleteMyBlockPackById(ctx *gin.Context, requestDto *capi.DeleteMyBlockPackByIdRequestDto)
	DeleteMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.DeleteMyBlockPacksByIdsRequestDto)
}

type BlockPackController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewBlockPackController(coreAdapter *coreadapters.CoreAdapter) BlockPackControllerInterface {
	return &BlockPackController{
		coreAdapter: coreAdapter,
	}
}

func (c *BlockPackController) GetMyBlockPackById(ctx *gin.Context, requestDto *capi.GetMyBlockPackByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlockPackByIdRequestDto,
		capi.GetMyBlockPackByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlockPackByIdOperation,
		"/core/v1/block-packs/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) GetMyBlockPackAndItsParentById(ctx *gin.Context, requestDto *capi.GetMyBlockPackAndItsParentByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlockPackAndItsParentByIdRequestDto,
		capi.GetMyBlockPackAndItsParentByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlockPackAndItsParentByIdOperation,
		"/core/v1/block-packs/get-and-parent-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) GetMyBlockPacksByParentSubShelfId(ctx *gin.Context, requestDto *capi.GetMyBlockPacksByParentSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlockPacksByParentSubShelfIdRequestDto,
		capi.GetMyBlockPacksByParentSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlockPacksByParentSubShelfIdOperation,
		"/core/v1/block-packs/get-by-parent-sub-shelf-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) GetAllMyBlockPacksByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMyBlockPacksByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetAllMyBlockPacksByRootShelfIdRequestDto,
		capi.GetAllMyBlockPacksByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyBlockPacksByRootShelfIdOperation,
		"/core/v1/block-packs/get-all-by-root-shelf-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) CreateBlockPack(ctx *gin.Context, requestDto *capi.CreateBlockPackRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateBlockPackRequestDto,
		capi.CreateBlockPackResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateBlockPackOperation,
		"/core/v1/block-packs/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *BlockPackController) CreateBlockPacks(ctx *gin.Context, requestDto *capi.CreateBlockPacksRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateBlockPacksRequestDto,
		capi.CreateBlockPacksResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateBlockPacksOperation,
		"/core/v1/block-packs/create-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *BlockPackController) UpdateMyBlockPackById(ctx *gin.Context, requestDto *capi.UpdateMyBlockPackByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyBlockPackByIdRequestDto,
		capi.UpdateMyBlockPackByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyBlockPackByIdOperation,
		"/core/v1/block-packs/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) UpdateMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.UpdateMyBlockPacksByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyBlockPacksByIdsRequestDto,
		capi.UpdateMyBlockPacksByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyBlockPacksByIdsOperation,
		"/core/v1/block-packs/update-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) MoveMyBlockPackByParentSubShelfId(ctx *gin.Context, requestDto *capi.MoveMyBlockPackByParentSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMyBlockPackByParentSubShelfIdRequestDto,
		capi.MoveMyBlockPackByParentSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMyBlockPackByParentSubShelfIdOperation,
		"/core/v1/block-packs/move",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) MoveMyBlockPacksByParentSubShelfId(ctx *gin.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMyBlockPacksByParentSubShelfIdRequestDto,
		capi.MoveMyBlockPacksByParentSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMyBlockPacksByParentSubShelfIdOperation,
		"/core/v1/block-packs/move-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) MoveMyBlockPacksByParentSubShelfIds(ctx *gin.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto,
		capi.MoveMyBlockPacksByParentSubShelfIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMyBlockPacksByParentSubShelfIdsOperation,
		"/core/v1/block-packs/move-many-by-parent-sub-shelves",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) RestoreMyBlockPackById(ctx *gin.Context, requestDto *capi.RestoreMyBlockPackByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyBlockPackByIdRequestDto,
		capi.RestoreMyBlockPackByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyBlockPackByIdOperation,
		"/core/v1/block-packs/restore",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) RestoreMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.RestoreMyBlockPacksByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyBlockPacksByIdsRequestDto,
		capi.RestoreMyBlockPacksByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyBlockPacksByIdsOperation,
		"/core/v1/block-packs/restore-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) DeleteMyBlockPackById(ctx *gin.Context, requestDto *capi.DeleteMyBlockPackByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyBlockPackByIdRequestDto,
		capi.DeleteMyBlockPackByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyBlockPackByIdOperation,
		"/core/v1/block-packs/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *BlockPackController) DeleteMyBlockPacksByIds(ctx *gin.Context, requestDto *capi.DeleteMyBlockPacksByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyBlockPacksByIdsRequestDto,
		capi.DeleteMyBlockPacksByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyBlockPacksByIdsOperation,
		"/core/v1/block-packs/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
