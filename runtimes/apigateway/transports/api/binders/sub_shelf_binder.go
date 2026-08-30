package binders

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/sub-shelves"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
)

type SubShelfBinderInterface interface {
	BindGetMySubShelfById(controllerFunc controllers.Func[*capi.GetMySubShelfByIdRequestDto]) gin.HandlerFunc
	BindGetMySubShelvesByPrevSubShelfId(controllerFunc controllers.Func[*capi.GetMySubShelvesByPrevSubShelfIdRequestDto]) gin.HandlerFunc
	BindGetAllMySubShelvesByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMySubShelvesByRootShelfIdRequestDto]) gin.HandlerFunc
	BindGetMySubShelvesAndItemsByPrevSubShelfId(controllerFunc controllers.Func[*capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto]) gin.HandlerFunc
	BindCreateSubShelfByRootShelfId(controllerFunc controllers.Func[*capi.CreateSubShelfByRootShelfIdRequestDto]) gin.HandlerFunc
	BindCreateSubShelvesByRootShelfIds(controllerFunc controllers.Func[*capi.CreateSubShelvesByRootShelfIdsRequestDto]) gin.HandlerFunc
	BindUpdateMySubShelfById(controllerFunc controllers.Func[*capi.UpdateMySubShelfByIdRequestDto]) gin.HandlerFunc
	BindUpdateMySubShelvesByIds(controllerFunc controllers.Func[*capi.UpdateMySubShelvesByIdsRequestDto]) gin.HandlerFunc
	BindMoveMySubShelfByRootShelfId(controllerFunc controllers.Func[*capi.MoveMySubShelfByRootShelfIdRequestDto]) gin.HandlerFunc
	BindMoveMySubShelvesByRootShelfId(controllerFunc controllers.Func[*capi.MoveMySubShelvesByRootShelfIdRequestDto]) gin.HandlerFunc
	BindMoveMySubShelvesByRootShelfIds(controllerFunc controllers.Func[*capi.MoveMySubShelvesByRootShelfIdsRequestDto]) gin.HandlerFunc
	BindRestoreMySubShelfById(controllerFunc controllers.Func[*capi.RestoreMySubShelfByIdRequestDto]) gin.HandlerFunc
	BindRestoreMySubShelvesByIds(controllerFunc controllers.Func[*capi.RestoreMySubShelvesByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMySubShelfById(controllerFunc controllers.Func[*capi.DeleteMySubShelfByIdRequestDto]) gin.HandlerFunc
	BindDeleteMySubShelvesByIds(controllerFunc controllers.Func[*capi.DeleteMySubShelvesByIdsRequestDto]) gin.HandlerFunc
}

type SubShelfBinder struct{}

func NewSubShelfBinder() SubShelfBinderInterface {
	return &SubShelfBinder{}
}

func (b *SubShelfBinder) BindGetMySubShelfById(controllerFunc controllers.Func[*capi.GetMySubShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMySubShelfByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			isDeleted, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &isDeleted
		}

		subShelfId, err := uuid.Parse(ctx.Param("sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.SubShelfId = subShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindGetMySubShelvesByPrevSubShelfId(controllerFunc controllers.Func[*capi.GetMySubShelvesByPrevSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMySubShelvesByPrevSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			areDeleted, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &areDeleted
		}

		prevSubShelfId, err := uuid.Parse(ctx.Param("prev-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.PrevSubShelfId = prevSubShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindGetAllMySubShelvesByRootShelfId(controllerFunc controllers.Func[*capi.GetAllMySubShelvesByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetAllMySubShelvesByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			areDeleted, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &areDeleted
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindGetMySubShelvesAndItemsByPrevSubShelfId(controllerFunc controllers.Func[*capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			areDeleted, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &areDeleted
		}

		prevSubShelfId, err := uuid.Parse(ctx.Param("prev-sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.PrevSubShelfId = prevSubShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindCreateSubShelfByRootShelfId(controllerFunc controllers.Func[*capi.CreateSubShelfByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.CreateSubShelfByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.RootShelfId = rootShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindCreateSubShelvesByRootShelfIds(controllerFunc controllers.Func[*capi.CreateSubShelvesByRootShelfIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.CreateSubShelvesByRootShelfIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindUpdateMySubShelfById(controllerFunc controllers.Func[*capi.UpdateMySubShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.UpdateMySubShelfByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		subShelfId, err := uuid.Parse(ctx.Param("sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.SubShelfId = subShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindUpdateMySubShelvesByIds(controllerFunc controllers.Func[*capi.UpdateMySubShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.UpdateMySubShelvesByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindMoveMySubShelfByRootShelfId(controllerFunc controllers.Func[*capi.MoveMySubShelfByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMySubShelfByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		subShelfId, err := uuid.Parse(ctx.Param("sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Body.SourceSubShelfId = subShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindMoveMySubShelvesByRootShelfId(controllerFunc controllers.Func[*capi.MoveMySubShelvesByRootShelfIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMySubShelvesByRootShelfIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindMoveMySubShelvesByRootShelfIds(controllerFunc controllers.Func[*capi.MoveMySubShelvesByRootShelfIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.MoveMySubShelvesByRootShelfIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindRestoreMySubShelfById(controllerFunc controllers.Func[*capi.RestoreMySubShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMySubShelfByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		subShelfId, err := uuid.Parse(ctx.Param("sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.SubShelfId = subShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindRestoreMySubShelvesByIds(controllerFunc controllers.Func[*capi.RestoreMySubShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.RestoreMySubShelvesByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindDeleteMySubShelfById(controllerFunc controllers.Func[*capi.DeleteMySubShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMySubShelfByIdRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		subShelfId, err := uuid.Parse(ctx.Param("sub-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.SubShelfId = subShelfId

		controllerFunc(ctx, &requestDto)
	}
}

func (b *SubShelfBinder) BindDeleteMySubShelvesByIds(controllerFunc controllers.Func[*capi.DeleteMySubShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestDto capi.DeleteMySubShelvesByIdsRequestDto

		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &requestDto)
	}
}
