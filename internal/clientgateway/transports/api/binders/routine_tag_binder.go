package binders

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tags"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type RoutineTagBinderInterface interface {
	BindGetMyRoutineTagById(controllerFunc controllers.Func[*capi.GetMyRoutineTagByIdRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutineTags(controllerFunc controllers.Func[*capi.GetAllMyRoutineTagsRequestDto]) gin.HandlerFunc
	BindCreateRoutineTag(controllerFunc controllers.Func[*capi.CreateRoutineTagRequestDto]) gin.HandlerFunc
	BindCreateRoutineTags(controllerFunc controllers.Func[*capi.CreateRoutineTagsRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineTagById(controllerFunc controllers.Func[*capi.UpdateMyRoutineTagByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineTagsByIds(controllerFunc controllers.Func[*capi.UpdateMyRoutineTagsByIdsRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTagById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTagByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTagsByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTagsByIdsRequestDto]) gin.HandlerFunc
}

type RoutineTagBinder struct{}

func NewRoutineTagBinder() RoutineTagBinderInterface {
	return &RoutineTagBinder{}
}

func (b *RoutineTagBinder) BindGetMyRoutineTagById(controllerFunc controllers.Func[*capi.GetMyRoutineTagByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRoutineTagByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			isDeleted, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTag").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.IsDeleted = &isDeleted
		}

		routineTagId, err := uuid.Parse(ctx.Param("routine-tag-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTag").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RoutineTagId = routineTagId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindGetAllMyRoutineTags(controllerFunc controllers.Func[*capi.GetAllMyRoutineTagsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetAllMyRoutineTagsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			areDeleted, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTag").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.AreDeleted = &areDeleted
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindCreateRoutineTag(controllerFunc controllers.Func[*capi.CreateRoutineTagRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineTagRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTag").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindCreateRoutineTags(controllerFunc controllers.Func[*capi.CreateRoutineTagsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineTagsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTag").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindUpdateMyRoutineTagById(controllerFunc controllers.Func[*capi.UpdateMyRoutineTagByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRoutineTagByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTag").WithOrigin(err), ctx)
			return
		}

		routineTagId, err := uuid.Parse(ctx.Param("routine-tag-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTag").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RoutineTagId = routineTagId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindUpdateMyRoutineTagsByIds(controllerFunc controllers.Func[*capi.UpdateMyRoutineTagsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRoutineTagsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTag").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindHardDeleteMyRoutineTagById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTagByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutineTagByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		routineTagId, err := uuid.Parse(ctx.Param("routine-tag-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTag").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RoutineTagId = routineTagId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTagBinder) BindHardDeleteMyRoutineTagsByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTagsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutineTagsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTag").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}
