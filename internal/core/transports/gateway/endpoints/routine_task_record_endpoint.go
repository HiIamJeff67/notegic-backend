package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-records"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
)

type RoutineTaskRecordEndpointInterface interface {
	GetAllMyRoutineTaskRecordsByRoutineTaskId(ctx *gin.Context)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskRecordStatusCount(ctx *gin.Context)
	VisualizeMyRoutineTaskRecordPurposeCount(ctx *gin.Context)
	VisualizeMyRoutineTaskRecordScheduledAtCount(ctx *gin.Context)
	VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx *gin.Context)
	VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx *gin.Context)

	/* ============================== GraphQL Methods ============================== */
	SearchRoutineTaskRecords(ctx *gin.Context)
}

type RoutineTaskRecordEndpoint struct {
	routineTaskRecordService routineservices.RoutineTaskRecordServiceInterface
}

func NewRoutineTaskRecordEndpoint(
	routineTaskRecordService routineservices.RoutineTaskRecordServiceInterface,
) RoutineTaskRecordEndpointInterface {
	return &RoutineTaskRecordEndpoint{
		routineTaskRecordService: routineTaskRecordService,
	}
}

func (t *RoutineTaskRecordEndpoint) GetAllMyRoutineTaskRecordsByRoutineTaskId(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.GetAllMyRoutineTaskRecordsByRoutineTaskId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

/* ============================== Visualization Methods ============================== */

func (t *RoutineTaskRecordEndpoint) VisualizeMyRoutineTaskRecordStatusCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.VisualizeMyRoutineTaskRecordStatusCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskRecordStatusCountResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RoutineTaskRecordEndpoint) VisualizeMyRoutineTaskRecordPurposeCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.VisualizeMyRoutineTaskRecordPurposeCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskRecordPurposeCountResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RoutineTaskRecordEndpoint) VisualizeMyRoutineTaskRecordScheduledAtCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.VisualizeMyRoutineTaskRecordScheduledAtCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RoutineTaskRecordEndpoint) VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RoutineTaskRecordEndpoint) VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskRecordService.VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}
