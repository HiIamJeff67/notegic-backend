package binders

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
)

type BlockBinderInterface interface {
	BindGetMyBlockById(controllerFunc controllers.Func[*capi.GetMyBlockByIdRequestDto]) gin.HandlerFunc
	BindGetMyBlocksByIds(controllerFunc controllers.Func[*capi.GetMyBlocksByIdsRequestDto]) gin.HandlerFunc
	BindGetMyBlocksByBlockPackId(controllerFunc controllers.Func[*capi.GetMyBlocksByBlockPackIdRequestDto]) gin.HandlerFunc
}

type BlockBinder struct{}

func NewBlockBinder() BlockBinderInterface {
	return &BlockBinder{}
}

func (b *BlockBinder) BindGetMyBlockById(controllerFunc controllers.Func[*capi.GetMyBlockByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyBlockByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		blockId, err := uuid.Parse(ctx.Param("block-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Block").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockId = blockId

		controllerFunc(ctx, requestDto)
	}
}

func (b *BlockBinder) BindGetMyBlocksByIds(controllerFunc controllers.Func[*capi.GetMyBlocksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyBlocksByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindQuery(&requestDto.Param); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Block").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *BlockBinder) BindGetMyBlocksByBlockPackId(controllerFunc controllers.Func[*capi.GetMyBlocksByBlockPackIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyBlocksByBlockPackIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		blockPackId, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Block").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = blockPackId

		controllerFunc(ctx, requestDto)
	}
}
