package controllers

import (
	"github.com/gin-gonic/gin"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-records"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type RoutineTaskRecordControllerInterface interface {
	GetAllMyRoutineTaskRecordsByRoutineTaskId(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskRecordStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto)
	VisualizeMyRoutineTaskRecordPurposeCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto)
	VisualizeMyRoutineTaskRecordScheduledAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto)
	VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto)
	VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto)
}

type RoutineTaskRecordController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRoutineTaskRecordController(coreAdapter *coreadapters.CoreAdapter) RoutineTaskRecordControllerInterface {
	return &RoutineTaskRecordController{
		coreAdapter: coreAdapter,
	}
}

func (c *RoutineTaskRecordController) GetAllMyRoutineTaskRecordsByRoutineTaskId(ctx *gin.Context, requestDto *capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto, capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdResponseDto](ctx, c.coreAdapter, requestDto, capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdOperation, "/core/v1/routine-task-records/get-all-by-routine-task-id")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

/* ============================== Visualization Methods ============================== */

func (c *RoutineTaskRecordController) VisualizeMyRoutineTaskRecordStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto, capi.VisualizeMyRoutineTaskRecordStatusCountResponseDto](ctx, c.coreAdapter, requestDto, capi.VisualizeMyRoutineTaskRecordStatusCountOperation, "/core/v1/routine-task-records/visualizations/status-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskRecordController) VisualizeMyRoutineTaskRecordPurposeCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto, capi.VisualizeMyRoutineTaskRecordPurposeCountResponseDto](ctx, c.coreAdapter, requestDto, capi.VisualizeMyRoutineTaskRecordPurposeCountOperation, "/core/v1/routine-task-records/visualizations/purpose-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskRecordController) VisualizeMyRoutineTaskRecordScheduledAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto, capi.VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto](ctx, c.coreAdapter, requestDto, capi.VisualizeMyRoutineTaskRecordScheduledAtCountOperation, "/core/v1/routine-task-records/visualizations/scheduled-at-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskRecordController) VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto, capi.VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto](ctx, c.coreAdapter, requestDto, capi.VisualizeMyRoutineTaskRecordActualStartedAtCountOperation, "/core/v1/routine-task-records/visualizations/actual-started-at-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *RoutineTaskRecordController) VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto, capi.VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto](ctx, c.coreAdapter, requestDto, capi.VisualizeMyRoutineTaskRecordActualEndedAtCountOperation, "/core/v1/routine-task-records/visualizations/actual-ended-at-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}
