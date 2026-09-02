package binders

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type RoutineTaskDependencyBinderInterface interface {
	BindGetRoutineTaskDependenciesByRoutineId(
		controllers.Func[*capi.GetRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindCreateRoutineTaskDependencyByRoutineId(
		controllers.Func[*capi.CreateRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindCreateRoutineTaskDependenciesByRoutineId(
		controllers.Func[*capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindUpdateRoutineTaskDependencyByRoutineId(
		controllers.Func[*capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindUpdateRoutineTaskDependenciesByRoutineId(
		controllers.Func[*capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindDeleteRoutineTaskDependencyByRoutineId(
		controllers.Func[*capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindDeleteRoutineTaskDependenciesByRoutineId(
		controllers.Func[*capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
}

type RoutineTaskDependencyBinder struct{}

func NewRoutineTaskDependencyBinder() RoutineTaskDependencyBinderInterface {
	return &RoutineTaskDependencyBinder{}
}

func (b *RoutineTaskDependencyBinder) BindGetRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*capi.GetRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskDependencyBinder) BindCreateRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*capi.CreateRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindCreateRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindUpdateRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindUpdateRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindDeleteRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindDeleteRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !bindRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependencyJSON(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func bindRoutineTaskDependencyRoutineId(ctx *gin.Context, routineId *uuid.UUID) bool {
	value, err := uuid.Parse(ctx.Param("routine-id"))
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTaskDependency").WithOrigin(err), ctx)
		return false
	}
	*routineId = value
	return true
}

func bindRoutineTaskDependencyJSON[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTaskDependency").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}
