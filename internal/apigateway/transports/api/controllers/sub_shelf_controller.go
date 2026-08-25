package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/sub-shelves"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/core/adapters"
)

type SubShelfControllerInterface interface {
	GetMySubShelfById(ctx *gin.Context, requestDto *capi.GetMySubShelfByIdRequestDto)
	GetMySubShelvesByPrevSubShelfId(ctx *gin.Context, requestDto *capi.GetMySubShelvesByPrevSubShelfIdRequestDto)
	GetAllMySubShelvesByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMySubShelvesByRootShelfIdRequestDto)
	GetMySubShelvesAndItemsByPrevSubShelfId(ctx *gin.Context, requestDto *capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto)
	CreateSubShelfByRootShelfId(ctx *gin.Context, requestDto *capi.CreateSubShelfByRootShelfIdRequestDto)
	CreateSubShelvesByRootShelfIds(ctx *gin.Context, requestDto *capi.CreateSubShelvesByRootShelfIdsRequestDto)
	UpdateMySubShelfById(ctx *gin.Context, requestDto *capi.UpdateMySubShelfByIdRequestDto)
	UpdateMySubShelvesByIds(ctx *gin.Context, requestDto *capi.UpdateMySubShelvesByIdsRequestDto)
	MoveMySubShelfByRootShelfId(ctx *gin.Context, requestDto *capi.MoveMySubShelfByRootShelfIdRequestDto)
	MoveMySubShelvesByRootShelfId(ctx *gin.Context, requestDto *capi.MoveMySubShelvesByRootShelfIdRequestDto)
	MoveMySubShelvesByRootShelfIds(ctx *gin.Context, requestDto *capi.MoveMySubShelvesByRootShelfIdsRequestDto)
	RestoreMySubShelfById(ctx *gin.Context, requestDto *capi.RestoreMySubShelfByIdRequestDto)
	RestoreMySubShelvesByIds(ctx *gin.Context, requestDto *capi.RestoreMySubShelvesByIdsRequestDto)
	DeleteMySubShelfById(ctx *gin.Context, requestDto *capi.DeleteMySubShelfByIdRequestDto)
	DeleteMySubShelvesByIds(ctx *gin.Context, requestDto *capi.DeleteMySubShelvesByIdsRequestDto)
}

type SubShelfController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewSubShelfController(coreAdapter *coreadapters.CoreAdapter) SubShelfControllerInterface {
	return &SubShelfController{
		coreAdapter: coreAdapter,
	}
}

func (c *SubShelfController) GetMySubShelfById(ctx *gin.Context, requestDto *capi.GetMySubShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMySubShelfByIdRequestDto,
		capi.GetMySubShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMySubShelfByIdOperation,
		"/core/v1/sub-shelves/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) GetMySubShelvesByPrevSubShelfId(ctx *gin.Context, requestDto *capi.GetMySubShelvesByPrevSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMySubShelvesByPrevSubShelfIdRequestDto,
		capi.GetMySubShelvesByPrevSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMySubShelvesByPrevSubShelfIdOperation,
		"/core/v1/sub-shelves/get-by-prev-sub-shelf-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) GetAllMySubShelvesByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMySubShelvesByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetAllMySubShelvesByRootShelfIdRequestDto,
		capi.GetAllMySubShelvesByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMySubShelvesByRootShelfIdOperation,
		"/core/v1/sub-shelves/get-all-by-root-shelf-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) GetMySubShelvesAndItemsByPrevSubShelfId(ctx *gin.Context, requestDto *capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto,
		capi.GetMySubShelvesAndItemsByPrevSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMySubShelvesAndItemsByPrevSubShelfIdOperation,
		"/core/v1/sub-shelves/get-and-items-by-prev-sub-shelf-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) CreateSubShelfByRootShelfId(ctx *gin.Context, requestDto *capi.CreateSubShelfByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateSubShelfByRootShelfIdRequestDto,
		capi.CreateSubShelfByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateSubShelfByRootShelfIdOperation,
		"/core/v1/sub-shelves/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *SubShelfController) CreateSubShelvesByRootShelfIds(ctx *gin.Context, requestDto *capi.CreateSubShelvesByRootShelfIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateSubShelvesByRootShelfIdsRequestDto,
		capi.CreateSubShelvesByRootShelfIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateSubShelvesByRootShelfIdsOperation,
		"/core/v1/sub-shelves/create-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *SubShelfController) UpdateMySubShelfById(ctx *gin.Context, requestDto *capi.UpdateMySubShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMySubShelfByIdRequestDto,
		capi.UpdateMySubShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMySubShelfByIdOperation,
		"/core/v1/sub-shelves/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) UpdateMySubShelvesByIds(ctx *gin.Context, requestDto *capi.UpdateMySubShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMySubShelvesByIdsRequestDto,
		capi.UpdateMySubShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMySubShelvesByIdsOperation,
		"/core/v1/sub-shelves/update-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) MoveMySubShelfByRootShelfId(ctx *gin.Context, requestDto *capi.MoveMySubShelfByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMySubShelfByRootShelfIdRequestDto,
		capi.MoveMySubShelfByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMySubShelfByRootShelfIdOperation,
		"/core/v1/sub-shelves/move",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) MoveMySubShelvesByRootShelfId(ctx *gin.Context, requestDto *capi.MoveMySubShelvesByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMySubShelvesByRootShelfIdRequestDto,
		capi.MoveMySubShelvesByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMySubShelvesByRootShelfIdOperation,
		"/core/v1/sub-shelves/move-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) MoveMySubShelvesByRootShelfIds(ctx *gin.Context, requestDto *capi.MoveMySubShelvesByRootShelfIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMySubShelvesByRootShelfIdsRequestDto,
		capi.MoveMySubShelvesByRootShelfIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMySubShelvesByRootShelfIdsOperation,
		"/core/v1/sub-shelves/move-many-by-root-shelves",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) RestoreMySubShelfById(ctx *gin.Context, requestDto *capi.RestoreMySubShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMySubShelfByIdRequestDto,
		capi.RestoreMySubShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMySubShelfByIdOperation,
		"/core/v1/sub-shelves/restore",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) RestoreMySubShelvesByIds(ctx *gin.Context, requestDto *capi.RestoreMySubShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMySubShelvesByIdsRequestDto,
		capi.RestoreMySubShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMySubShelvesByIdsOperation,
		"/core/v1/sub-shelves/restore-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) DeleteMySubShelfById(ctx *gin.Context, requestDto *capi.DeleteMySubShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMySubShelfByIdRequestDto,
		capi.DeleteMySubShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMySubShelfByIdOperation,
		"/core/v1/sub-shelves/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *SubShelfController) DeleteMySubShelvesByIds(ctx *gin.Context, requestDto *capi.DeleteMySubShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMySubShelvesByIdsRequestDto,
		capi.DeleteMySubShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMySubShelvesByIdsOperation,
		"/core/v1/sub-shelves/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
