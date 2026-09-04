package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
)

type RoutineTaskEndpointInterface interface {
	GetMyRoutineTaskById(ctx *gin.Context)
	GetMyRoutineTasksByRoutineId(ctx *gin.Context)
	GetMyRoutineTasksByRoutineIds(ctx *gin.Context)
	GetAllMyRoutineTasks(ctx *gin.Context)
	CreateRoutineTaskByRoutineId(ctx *gin.Context)
	UpdateMyRoutineTaskById(ctx *gin.Context)
	HardDeleteMyRoutineTaskById(ctx *gin.Context)
	HardDeleteMyRoutineTasksByIds(ctx *gin.Context)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskPurposeCount(ctx *gin.Context)

	/* ============================== GraphQL Methods ============================== */
	SearchRoutineTasks(ctx *gin.Context)
}

type RoutineTaskEndpoint struct {
	routineTaskService routineservices.RoutineTaskServiceInterface
}

func NewRoutineTaskEndpoint(routineTaskService routineservices.RoutineTaskServiceInterface) RoutineTaskEndpointInterface {
	return &RoutineTaskEndpoint{routineTaskService: routineTaskService}
}

func (t *RoutineTaskEndpoint) GetMyRoutineTaskById(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRoutineTaskByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.GetMyRoutineTaskById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRoutineTaskByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) GetMyRoutineTasksByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRoutineTasksByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.GetMyRoutineTasksByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRoutineTasksByRoutineIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) GetMyRoutineTasksByRoutineIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRoutineTasksByRoutineIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.GetMyRoutineTasksByRoutineIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRoutineTasksByRoutineIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) GetAllMyRoutineTasks(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetAllMyRoutineTasksRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.GetAllMyRoutineTasks(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetAllMyRoutineTasksResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) CreateRoutineTaskByRoutineId(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRoutineTaskByRoutineIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.CreateRoutineTaskByRoutineId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateRoutineTaskByRoutineIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) UpdateMyRoutineTaskById(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRoutineTaskByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.UpdateMyRoutineTaskById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRoutineTaskByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) HardDeleteMyRoutineTaskById(ctx *gin.Context) {
	request := &cgateway.Request[capi.HardDeleteMyRoutineTaskByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.HardDeleteMyRoutineTaskById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.HardDeleteMyRoutineTaskByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) HardDeleteMyRoutineTasksByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.HardDeleteMyRoutineTasksByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.HardDeleteMyRoutineTasksByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.HardDeleteMyRoutineTasksByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineTaskEndpoint) VisualizeMyRoutineTaskPurposeCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineTaskPurposeCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineTaskService.VisualizeMyRoutineTaskPurposeCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineTaskPurposeCountResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}
