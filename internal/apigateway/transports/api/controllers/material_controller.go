package controllers

import (
	"github.com/gin-gonic/gin"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/materials"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/core/adapters"
)

type MaterialControllerInterface interface {
	GetMyMaterialById(ctx *gin.Context, requestDto *capi.GetMyMaterialByIdRequestDto)
	GetMyMaterialAndItsParentById(ctx *gin.Context, requestDto *capi.GetMyMaterialAndItsParentByIdRequestDto)
	GetMyMaterialsByParentSubShelfId(ctx *gin.Context, requestDto *capi.GetMyMaterialsByParentSubShelfIdRequestDto)
	GetAllMyMaterialsByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMyMaterialsByRootShelfIdRequestDto)
	CreateMyMaterial(ctx *gin.Context, requestDto *capi.CreateMyMaterialRequestDto)
	UpdateMyMaterialById(ctx *gin.Context, requestDto *capi.UpdateMyMaterialByIdRequestDto)
	SaveMyMaterialById(ctx *gin.Context, requestDto *capi.SaveMyMaterialByIdRequestDto)
	MoveMyMaterialById(ctx *gin.Context, requestDto *capi.MoveMyMaterialByIdRequestDto)
	MoveMyMaterialsByIds(ctx *gin.Context, requestDto *capi.MoveMyMaterialsByIdsRequestDto)
	RestoreMyMaterialById(ctx *gin.Context, requestDto *capi.RestoreMyMaterialByIdRequestDto)
	RestoreMyMaterialsByIds(ctx *gin.Context, requestDto *capi.RestoreMyMaterialsByIdsRequestDto)
	DeleteMyMaterialById(ctx *gin.Context, requestDto *capi.DeleteMyMaterialByIdRequestDto)
	DeleteMyMaterialsByIds(ctx *gin.Context, requestDto *capi.DeleteMyMaterialsByIdsRequestDto)
}

type MaterialController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewMaterialController(coreAdapter *coreadapters.CoreAdapter) MaterialControllerInterface {
	return &MaterialController{
		coreAdapter: coreAdapter,
	}
}

func (c *MaterialController) GetMyMaterialById(ctx *gin.Context, requestDto *capi.GetMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyMaterialByIdRequestDto,
		capi.GetMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyMaterialByIdOperation,
		"/core/v1/materials/get-by-id",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) GetMyMaterialAndItsParentById(ctx *gin.Context, requestDto *capi.GetMyMaterialAndItsParentByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyMaterialAndItsParentByIdRequestDto,
		capi.GetMyMaterialAndItsParentByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyMaterialAndItsParentByIdOperation,
		"/core/v1/materials/get-and-parent-by-id",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) GetMyMaterialsByParentSubShelfId(ctx *gin.Context, requestDto *capi.GetMyMaterialsByParentSubShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyMaterialsByParentSubShelfIdRequestDto,
		capi.GetMyMaterialsByParentSubShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyMaterialsByParentSubShelfIdOperation,
		"/core/v1/materials/get-by-parent-sub-shelf-id",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) GetAllMyMaterialsByRootShelfId(ctx *gin.Context, requestDto *capi.GetAllMyMaterialsByRootShelfIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetAllMyMaterialsByRootShelfIdRequestDto,
		capi.GetAllMyMaterialsByRootShelfIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyMaterialsByRootShelfIdOperation,
		"/core/v1/materials/get-all-by-root-shelf-id",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) CreateMyMaterial(ctx *gin.Context, requestDto *capi.CreateMyMaterialRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateMyMaterialRequestDto,
		capi.CreateMyMaterialResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateMyMaterialOperation,
		"/core/v1/materials/create",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *MaterialController) UpdateMyMaterialById(ctx *gin.Context, requestDto *capi.UpdateMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyMaterialByIdRequestDto,
		capi.UpdateMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyMaterialByIdOperation,
		"/core/v1/materials/update",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) SaveMyMaterialById(ctx *gin.Context, requestDto *capi.SaveMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.SaveMyMaterialByIdRequestDto,
		capi.SaveMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.SaveMyMaterialByIdOperation,
		"/core/v1/materials/save",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) MoveMyMaterialById(ctx *gin.Context, requestDto *capi.MoveMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMyMaterialByIdRequestDto,
		capi.MoveMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMyMaterialByIdOperation,
		"/core/v1/materials/move",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) MoveMyMaterialsByIds(ctx *gin.Context, requestDto *capi.MoveMyMaterialsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.MoveMyMaterialsByIdsRequestDto,
		capi.MoveMyMaterialsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.MoveMyMaterialsByIdsOperation,
		"/core/v1/materials/move-many",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) RestoreMyMaterialById(ctx *gin.Context, requestDto *capi.RestoreMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyMaterialByIdRequestDto,
		capi.RestoreMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyMaterialByIdOperation,
		"/core/v1/materials/restore",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) RestoreMyMaterialsByIds(ctx *gin.Context, requestDto *capi.RestoreMyMaterialsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyMaterialsByIdsRequestDto,
		capi.RestoreMyMaterialsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyMaterialsByIdsOperation,
		"/core/v1/materials/restore-many",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) DeleteMyMaterialById(ctx *gin.Context, requestDto *capi.DeleteMyMaterialByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyMaterialByIdRequestDto,
		capi.DeleteMyMaterialByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyMaterialByIdOperation,
		"/core/v1/materials/delete",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *MaterialController) DeleteMyMaterialsByIds(ctx *gin.Context, requestDto *capi.DeleteMyMaterialsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyMaterialsByIdsRequestDto,
		capi.DeleteMyMaterialsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyMaterialsByIdsOperation,
		"/core/v1/materials/delete-many",
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
