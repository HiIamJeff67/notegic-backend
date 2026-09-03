package controllers

import (
	"net/http"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type BlockControllerInterface interface {
	GetMyBlockById(ctx *gin.Context, requestDto *capi.GetMyBlockByIdRequestDto)
	GetMyBlocksByIds(ctx *gin.Context, requestDto *capi.GetMyBlocksByIdsRequestDto)
	GetMyBlocksByBlockPackId(ctx *gin.Context, requestDto *capi.GetMyBlocksByBlockPackIdRequestDto)
}

type BlockController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewBlockController(coreAdapter *coreadapters.CoreAdapter) BlockControllerInterface {
	return &BlockController{
		coreAdapter: coreAdapter,
	}
}

func (c *BlockController) GetMyBlockById(ctx *gin.Context, requestDto *capi.GetMyBlockByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlockByIdRequestDto,
		capi.GetMyBlockByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlockByIdOperation,
		"/core/v1/blocks/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *BlockController) GetMyBlocksByIds(ctx *gin.Context, requestDto *capi.GetMyBlocksByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlocksByIdsRequestDto,
		capi.GetMyBlocksByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlocksByIdsOperation,
		"/core/v1/blocks/get-by-ids",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *BlockController) GetMyBlocksByBlockPackId(ctx *gin.Context, requestDto *capi.GetMyBlocksByBlockPackIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyBlocksByBlockPackIdRequestDto,
		capi.GetMyBlocksByBlockPackIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyBlocksByBlockPackIdOperation,
		"/core/v1/blocks/get-by-block-pack-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}
