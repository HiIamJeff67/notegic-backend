package controllers

import (
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
	PauseMyRoutineTaskById(ctx *gin.Context, requestDto *capi.PauseMyRoutineTaskByIdRequestDto)
	ResumeMyRoutineTaskById(ctx *gin.Context, requestDto *capi.ResumeMyRoutineTaskByIdRequestDto)
	HardDeleteMyRoutineTaskById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTaskByIdRequestDto)
	HardDeleteMyRoutineTasksByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskStatusCountRequestDto)
	VisualizeMyRoutineTaskPurposeCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto)
	VisualizeMyRoutineTaskScheduledAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskScheduledAtCountRequestDto)
	VisualizeMyRoutineTaskActualStartedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskActualStartedAtCountRequestDto)
	VisualizeMyRoutineTaskActualEndedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskActualEndedAtCountRequestDto)
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

	writeClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
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

	writeCreatedClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) PauseMyRoutineTaskById(ctx *gin.Context, requestDto *capi.PauseMyRoutineTaskByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.PauseMyRoutineTaskByIdRequestDto, capi.PauseMyRoutineTaskByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.PauseMyRoutineTaskByIdOperation,
		"/core/v1/routine-tasks/pause",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) ResumeMyRoutineTaskById(ctx *gin.Context, requestDto *capi.ResumeMyRoutineTaskByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.ResumeMyRoutineTaskByIdRequestDto, capi.ResumeMyRoutineTaskByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.ResumeMyRoutineTaskByIdOperation,
		"/core/v1/routine-tasks/resume",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) VisualizeMyRoutineTaskStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskStatusCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskStatusCountRequestDto, capi.VisualizeMyRoutineTaskStatusCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineTaskStatusCountOperation,
		"/core/v1/routine-tasks/visualize-status-count",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
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

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) VisualizeMyRoutineTaskScheduledAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskScheduledAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskScheduledAtCountRequestDto, capi.VisualizeMyRoutineTaskScheduledAtCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineTaskScheduledAtCountOperation,
		"/core/v1/routine-tasks/visualize-scheduled-at-count",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) VisualizeMyRoutineTaskActualStartedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskActualStartedAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskActualStartedAtCountRequestDto, capi.VisualizeMyRoutineTaskActualStartedAtCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineTaskActualStartedAtCountOperation,
		"/core/v1/routine-tasks/visualize-actual-started-at-count",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskController) VisualizeMyRoutineTaskActualEndedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskActualEndedAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskActualEndedAtCountRequestDto, capi.VisualizeMyRoutineTaskActualEndedAtCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineTaskActualEndedAtCountOperation,
		"/core/v1/routine-tasks/visualize-actual-ended-at-count",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
