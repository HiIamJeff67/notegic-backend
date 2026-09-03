package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type RoutineTaskDependencyControllerInterface interface {
	GetRoutineTaskDependenciesByRoutineId(*gin.Context, *capi.GetRoutineTaskDependenciesByRoutineIdRequestDto)
	CreateRoutineTaskDependencyByRoutineId(*gin.Context, *capi.CreateRoutineTaskDependencyByRoutineIdRequestDto)
	CreateRoutineTaskDependenciesByRoutineId(*gin.Context, *capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto)
	UpdateRoutineTaskDependencyByRoutineId(*gin.Context, *capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto)
	UpdateRoutineTaskDependenciesByRoutineId(*gin.Context, *capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto)
	DeleteRoutineTaskDependencyByRoutineId(*gin.Context, *capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto)
	DeleteRoutineTaskDependenciesByRoutineId(*gin.Context, *capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto)
}

type RoutineTaskDependencyController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRoutineTaskDependencyController(coreAdapter *coreadapters.CoreAdapter) RoutineTaskDependencyControllerInterface {
	return &RoutineTaskDependencyController{
		coreAdapter: coreAdapter,
	}
}

func (c *RoutineTaskDependencyController) GetRoutineTaskDependenciesByRoutineId(
	ctx *gin.Context,
	requestDto *capi.GetRoutineTaskDependenciesByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.GetRoutineTaskDependenciesByRoutineIdRequestDto,
		capi.GetRoutineTaskDependenciesByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetRoutineTaskDependenciesByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/get-by-routine-id",
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

func (c *RoutineTaskDependencyController) CreateRoutineTaskDependencyByRoutineId(
	ctx *gin.Context,
	requestDto *capi.CreateRoutineTaskDependencyByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRoutineTaskDependencyByRoutineIdRequestDto,
		capi.CreateRoutineTaskDependencyByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineTaskDependencyByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/create-by-routine-id",
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

func (c *RoutineTaskDependencyController) CreateRoutineTaskDependenciesByRoutineId(
	ctx *gin.Context,
	requestDto *capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto,
		capi.CreateRoutineTaskDependenciesByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineTaskDependenciesByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/create-many-by-routine-id",
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

func (c *RoutineTaskDependencyController) UpdateRoutineTaskDependencyByRoutineId(
	ctx *gin.Context,
	requestDto *capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto,
		capi.UpdateRoutineTaskDependencyByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateRoutineTaskDependencyByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/update-by-routine-id",
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

func (c *RoutineTaskDependencyController) UpdateRoutineTaskDependenciesByRoutineId(
	ctx *gin.Context,
	requestDto *capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto,
		capi.UpdateRoutineTaskDependenciesByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateRoutineTaskDependenciesByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/update-many-by-routine-id",
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

func (c *RoutineTaskDependencyController) DeleteRoutineTaskDependencyByRoutineId(
	ctx *gin.Context,
	requestDto *capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto,
		capi.DeleteRoutineTaskDependencyByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteRoutineTaskDependencyByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/delete-by-routine-id",
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

func (c *RoutineTaskDependencyController) DeleteRoutineTaskDependenciesByRoutineId(
	ctx *gin.Context,
	requestDto *capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto,
		capi.DeleteRoutineTaskDependenciesByRoutineIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteRoutineTaskDependenciesByRoutineIdOperation,
		"/core/v1/routine-task-dependencies/delete-many-by-routine-id",
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
