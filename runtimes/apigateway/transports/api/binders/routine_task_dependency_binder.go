package binders

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
)

type RoutineTaskDependencyBinderInterface interface {
	BindGetRoutineTaskDependenciesByRoutineId(
		controllers.Func[*apicontract.GetRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindCreateRoutineTaskDependencyByRoutineId(
		controllers.Func[*apicontract.CreateRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindCreateRoutineTaskDependenciesByRoutineId(
		controllers.Func[*apicontract.CreateRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindUpdateRoutineTaskDependencyByRoutineId(
		controllers.Func[*apicontract.UpdateRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindUpdateRoutineTaskDependenciesByRoutineId(
		controllers.Func[*apicontract.UpdateRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindDeleteRoutineTaskDependencyByRoutineId(
		controllers.Func[*apicontract.DeleteRoutineTaskDependencyByRoutineIdRequestDto],
	) gin.HandlerFunc
	BindDeleteRoutineTaskDependenciesByRoutineId(
		controllers.Func[*apicontract.DeleteRoutineTaskDependenciesByRoutineIdRequestDto],
	) gin.HandlerFunc
}

type RoutineTaskDependencyBinder struct{}

func NewRoutineTaskDependencyBinder() RoutineTaskDependencyBinderInterface {
	return &RoutineTaskDependencyBinder{}
}

func bindRoutineTaskDependency[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTaskDependency").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}

func parseRoutineTaskDependencyRoutineId(ctx *gin.Context, routineId *uuid.UUID) bool {
	value, err := uuid.Parse(ctx.Param("routine-id"))
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTaskDependency").WithOrigin(err), ctx)
		return false
	}
	*routineId = value
	return true
}

func (b *RoutineTaskDependencyBinder) BindGetRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*apicontract.GetRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskDependencyBinder) BindCreateRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*apicontract.CreateRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.CreateRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindCreateRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*apicontract.CreateRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.CreateRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindUpdateRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*apicontract.UpdateRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.UpdateRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindUpdateRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*apicontract.UpdateRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.UpdateRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindDeleteRoutineTaskDependencyByRoutineId(
	controllerFunc controllers.Func[*apicontract.DeleteRoutineTaskDependencyByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.DeleteRoutineTaskDependencyByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}

func (b *RoutineTaskDependencyBinder) BindDeleteRoutineTaskDependenciesByRoutineId(
	controllerFunc controllers.Func[*apicontract.DeleteRoutineTaskDependenciesByRoutineIdRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.DeleteRoutineTaskDependenciesByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if !parseRoutineTaskDependencyRoutineId(ctx, &requestDto.Param.RoutineId) {
			return
		}
		bindRoutineTaskDependency(
			ctx,
			requestDto,
			&requestDto.Body,
			controllerFunc,
		)
	}
}
