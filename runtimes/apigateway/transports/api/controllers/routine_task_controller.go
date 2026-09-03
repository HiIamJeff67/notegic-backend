package controllers

import (
	"net/http"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type RoutineTaskControllerInterface interface {
	GetMyRoutineTaskById(ctx *gin.Context, requestDto *capi.GetMyRoutineTaskByIdRequestDto)
	GetAllMyRoutineTasksByRoutineIds(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTasksByRoutineIdsRequestDto)
	GetAllMyRoutineTasks(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTasksRequestDto)
	CreateRoutineTaskByRoutineId(ctx *gin.Context, requestDto *capi.CreateRoutineTaskByRoutineIdRequestDto)
	UpdateMyRoutineTaskById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTaskByIdRequestDto)
	HardDeleteMyRoutineTaskById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTaskByIdRequestDto)
	HardDeleteMyRoutineTasksByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskPurposeCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto)
}

type RoutineTaskController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRoutineTaskController(coreAdapter *coreadapters.CoreAdapter) RoutineTaskControllerInterface {
	return &RoutineTaskController{coreAdapter: coreAdapter}
}

func (c *RoutineTaskController) GetMyRoutineTaskById(ctx *gin.Context, requestDto *capi.GetMyRoutineTaskByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetMyRoutineTaskByIdRequestDto, capi.GetMyRoutineTaskByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyRoutineTaskByIdOperation,
		"/core/v1/routine-tasks/get-by-id",
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

func (c *RoutineTaskController) GetAllMyRoutineTasksByRoutineIds(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTasksByRoutineIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetAllMyRoutineTasksByRoutineIdsRequestDto, capi.GetAllMyRoutineTasksByRoutineIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyRoutineTasksByRoutineIdsOperation,
		"/core/v1/routine-tasks/get-all-by-routine-ids",
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

func (c *RoutineTaskController) GetAllMyRoutineTasks(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTasksRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetAllMyRoutineTasksRequestDto, capi.GetAllMyRoutineTasksResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyRoutineTasksOperation,
		"/core/v1/routine-tasks/get-all",
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

func (c *RoutineTaskController) CreateRoutineTaskByRoutineId(ctx *gin.Context, requestDto *capi.CreateRoutineTaskByRoutineIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.CreateRoutineTaskByRoutineIdRequestDto, capi.CreateRoutineTaskByRoutineIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineTaskByRoutineIdOperation,
		"/core/v1/routine-tasks/create-by-routine-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusCreated,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *RoutineTaskController) UpdateMyRoutineTaskById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineTaskByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.UpdateMyRoutineTaskByIdRequestDto, capi.UpdateMyRoutineTaskByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRoutineTaskByIdOperation,
		"/core/v1/routine-tasks/update",
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

func (c *RoutineTaskController) HardDeleteMyRoutineTaskById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTaskByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.HardDeleteMyRoutineTaskByIdRequestDto, capi.HardDeleteMyRoutineTaskByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutineTaskByIdOperation,
		"/core/v1/routine-tasks/hard-delete",
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

func (c *RoutineTaskController) HardDeleteMyRoutineTasksByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.HardDeleteMyRoutineTasksByIdsRequestDto, capi.HardDeleteMyRoutineTasksByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutineTasksByIdsOperation,
		"/core/v1/routine-tasks/hard-delete-many",
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

func (c *RoutineTaskController) VisualizeMyRoutineTaskPurposeCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskPurposeCountRequestDto, capi.VisualizeMyRoutineTaskPurposeCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineTaskPurposeCountOperation,
		"/core/v1/routine-tasks/visualize-purpose-count",
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
