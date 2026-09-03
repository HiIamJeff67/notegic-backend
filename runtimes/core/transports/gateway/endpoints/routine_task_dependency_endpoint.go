package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
)

type RoutineTaskDependencyEndpointInterface interface {
	GetRoutineTaskDependenciesByRoutineId(ctx *gin.Context)
	CreateRoutineTaskDependencyByRoutineId(ctx *gin.Context)
	CreateRoutineTaskDependenciesByRoutineId(ctx *gin.Context)
	UpdateRoutineTaskDependencyByRoutineId(ctx *gin.Context)
	UpdateRoutineTaskDependenciesByRoutineId(ctx *gin.Context)
	DeleteRoutineTaskDependencyByRoutineId(ctx *gin.Context)
	DeleteRoutineTaskDependenciesByRoutineId(ctx *gin.Context)
}

type RoutineTaskDependencyEndpoint struct {
	routineTaskDependencyService routineservices.RoutineTaskDependencyServiceInterface
}

func NewRoutineTaskDependencyEndpoint(
	routineTaskDependencyService routineservices.RoutineTaskDependencyServiceInterface,
) RoutineTaskDependencyEndpointInterface {
	return &RoutineTaskDependencyEndpoint{
		routineTaskDependencyService: routineTaskDependencyService,
	}
}

func writeRoutineTaskDependencySuccess[D any](
	ctx *gin.Context,
	requestId string,
	data D,
) {
	ctx.JSON(
		http.StatusOK,
		cgateway.Response[D]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   requestId,
				RespondedAt: time.Now(),
			},
			Data: data,
		},
	)
}

func writeRoutineTaskDependencyException(ctx *gin.Context, requestId string, exception *cexceptions.Exception) {
	publicException := exception.ToPublic()
	ctx.JSON(
		publicException.HTTPStatusCode(),
		cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   requestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		},
	)
}

func (t *RoutineTaskDependencyEndpoint) GetRoutineTaskDependenciesByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetRoutineTaskDependenciesByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.GetRoutineTaskDependenciesByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) CreateRoutineTaskDependencyByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRoutineTaskDependencyByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.CreateRoutineTaskDependencyByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) CreateRoutineTaskDependenciesByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.CreateRoutineTaskDependenciesByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) UpdateRoutineTaskDependencyByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.UpdateRoutineTaskDependencyByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) UpdateRoutineTaskDependenciesByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.UpdateRoutineTaskDependenciesByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) DeleteRoutineTaskDependencyByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.DeleteRoutineTaskDependencyByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}

func (t *RoutineTaskDependencyEndpoint) DeleteRoutineTaskDependenciesByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseDto, exception := t.routineTaskDependencyService.DeleteRoutineTaskDependenciesByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		writeRoutineTaskDependencyException(ctx, request.Metadata.RequestId, exception)
		return
	}
	writeRoutineTaskDependencySuccess(
		ctx,
		request.Metadata.RequestId,
		*responseDto,
	)
}
