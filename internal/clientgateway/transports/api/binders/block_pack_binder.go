package binders

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type BlockPackBinderInterface interface {
	BindGetMyBlockPackById(controllerFunc controllers.Func[*capi.GetMyBlockPackByIdRequestDto]) gin.HandlerFunc
	BindGetMyBlockPackAndItsParentById(controllerFunc controllers.Func[*capi.GetMyBlockPackAndItsParentByIdRequestDto]) gin.HandlerFunc
	BindGetMyBlockPacksByParentSubShelfId(controllerFunc controllers.Func[*capi.GetMyBlockPacksByParentSubShelfIdRequestDto]) gin.HandlerFunc
	BindGetAllMyBlockPacksByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMyBlockPacksByRootShelfIdRequestDto]) gin.HandlerFunc
	BindCreateBlockPack(controllerFunc controllers.Func[*capi.CreateBlockPackRequestDto]) gin.HandlerFunc
	BindCreateBlockPacks(controllerFunc controllers.Func[*capi.CreateBlockPacksRequestDto]) gin.HandlerFunc
	BindUpdateMyBlockPackById(controllerFunc controllers.Func[*capi.UpdateMyBlockPackByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyBlockPacksByIds(controllerFunc controllers.Func[*capi.UpdateMyBlockPacksByIdsRequestDto]) gin.HandlerFunc
	BindMoveMyBlockPackByParentSubShelfId(controllerFunc controllers.Func[*capi.MoveMyBlockPackByParentSubShelfIdRequestDto]) gin.HandlerFunc
	BindMoveMyBlockPacksByParentSubShelfId(controllerFunc controllers.Func[*capi.MoveMyBlockPacksByParentSubShelfIdRequestDto]) gin.HandlerFunc
	BindMoveMyBlockPacksByParentSubShelfIds(controllerFunc controllers.Func[*capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyBlockPackById(controllerFunc controllers.Func[*capi.RestoreMyBlockPackByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyBlockPacksByIds(controllerFunc controllers.Func[*capi.RestoreMyBlockPacksByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyBlockPackById(controllerFunc controllers.Func[*capi.DeleteMyBlockPackByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyBlockPacksByIds(controllerFunc controllers.Func[*capi.DeleteMyBlockPacksByIdsRequestDto]) gin.HandlerFunc
}

type BlockPackBinder struct{}

func NewBlockPackBinder() BlockPackBinderInterface {
	return &BlockPackBinder{}
}

func (b *BlockPackBinder) BindGetMyBlockPackById(controllerFunc controllers.Func[*capi.GetMyBlockPackByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyBlockPackByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			value, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindGetMyBlockPackAndItsParentById(controllerFunc controllers.Func[*capi.GetMyBlockPackAndItsParentByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyBlockPackAndItsParentByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			value, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindGetMyBlockPacksByParentSubShelfId(controllerFunc controllers.Func[*capi.GetMyBlockPacksByParentSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMyBlockPacksByParentSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			value, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("parent-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.ParentSubShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindGetAllMyBlockPacksByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMyBlockPacksByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetAllMyBlockPacksByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			value, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &value
		}

		value, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindCreateBlockPack(controllerFunc controllers.Func[*capi.CreateBlockPackRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.CreateBlockPackRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		value, err := uuid.Parse(ctx.Param("parent-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.ParentSubShelfId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindCreateBlockPacks(controllerFunc controllers.Func[*capi.CreateBlockPacksRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.CreateBlockPacksRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindUpdateMyBlockPackById(controllerFunc controllers.Func[*capi.UpdateMyBlockPackByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.UpdateMyBlockPackByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindUpdateMyBlockPacksByIds(controllerFunc controllers.Func[*capi.UpdateMyBlockPacksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.UpdateMyBlockPacksByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindMoveMyBlockPackByParentSubShelfId(controllerFunc controllers.Func[*capi.MoveMyBlockPackByParentSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMyBlockPackByParentSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindMoveMyBlockPacksByParentSubShelfId(controllerFunc controllers.Func[*capi.MoveMyBlockPacksByParentSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMyBlockPacksByParentSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindMoveMyBlockPacksByParentSubShelfIds(controllerFunc controllers.Func[*capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindRestoreMyBlockPackById(controllerFunc controllers.Func[*capi.RestoreMyBlockPackByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMyBlockPackByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindRestoreMyBlockPacksByIds(controllerFunc controllers.Func[*capi.RestoreMyBlockPacksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMyBlockPacksByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindDeleteMyBlockPackById(controllerFunc controllers.Func[*capi.DeleteMyBlockPackByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMyBlockPackByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		value, err := uuid.Parse(ctx.Param("block-pack-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("BlockPack").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.BlockPackId = value

		controllerFunc(ctx, &requestDto)
	}
}

func (b *BlockPackBinder) BindDeleteMyBlockPacksByIds(controllerFunc controllers.Func[*capi.DeleteMyBlockPacksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMyBlockPacksByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("BlockPack").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}
