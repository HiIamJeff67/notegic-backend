package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tags"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type RoutineTagControllerInterface interface {
	GetMyRoutineTagById(ctx *gin.Context, requestDto *capi.GetMyRoutineTagByIdRequestDto)
	GetAllMyRoutineTags(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTagsRequestDto)
	CreateRoutineTag(ctx *gin.Context, requestDto *capi.CreateRoutineTagRequestDto)
	CreateRoutineTags(ctx *gin.Context, requestDto *capi.CreateRoutineTagsRequestDto)
	UpdateMyRoutineTagById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTagByIdRequestDto)
	UpdateMyRoutineTagsByIds(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTagsByIdsRequestDto)
	HardDeleteMyRoutineTagById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTagByIdRequestDto)
	HardDeleteMyRoutineTagsByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTagsByIdsRequestDto)
}

type RoutineTagController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRoutineTagController(coreAdapter *coreadapters.CoreAdapter) RoutineTagControllerInterface {
	return &RoutineTagController{
		coreAdapter: coreAdapter,
	}
}

func (c *RoutineTagController) GetMyRoutineTagById(ctx *gin.Context, requestDto *capi.GetMyRoutineTagByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyRoutineTagByIdRequestDto,
		capi.GetMyRoutineTagByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyRoutineTagByIdOperation,
		"/core/v1/routine-tags/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) GetAllMyRoutineTags(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTagsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetAllMyRoutineTagsRequestDto,
		capi.GetAllMyRoutineTagsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyRoutineTagsOperation,
		"/core/v1/routine-tags/get-all",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) CreateRoutineTag(ctx *gin.Context, requestDto *capi.CreateRoutineTagRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRoutineTagRequestDto,
		capi.CreateRoutineTagResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineTagOperation,
		"/core/v1/routine-tags/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) CreateRoutineTags(ctx *gin.Context, requestDto *capi.CreateRoutineTagsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRoutineTagsRequestDto,
		capi.CreateRoutineTagsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineTagsOperation,
		"/core/v1/routine-tags/create-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) UpdateMyRoutineTagById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTagByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyRoutineTagByIdRequestDto,
		capi.UpdateMyRoutineTagByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRoutineTagByIdOperation,
		"/core/v1/routine-tags/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) UpdateMyRoutineTagsByIds(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTagsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyRoutineTagsByIdsRequestDto,
		capi.UpdateMyRoutineTagsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRoutineTagsByIdsOperation,
		"/core/v1/routine-tags/update-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) HardDeleteMyRoutineTagById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTagByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.HardDeleteMyRoutineTagByIdRequestDto,
		capi.HardDeleteMyRoutineTagByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutineTagByIdOperation,
		"/core/v1/routine-tags/hard-delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTagController) HardDeleteMyRoutineTagsByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTagsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.HardDeleteMyRoutineTagsByIdsRequestDto,
		capi.HardDeleteMyRoutineTagsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutineTagsByIdsOperation,
		"/core/v1/routine-tags/hard-delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
