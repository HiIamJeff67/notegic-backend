package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
)

type RoutineEndpointInterface interface {
	GetMyRoutineById(ctx *gin.Context)
	GetMyRoutinesByStationId(ctx *gin.Context)
	GetAllMyRoutinesByTimeRange(ctx *gin.Context)
	CreateRoutineByStationId(ctx *gin.Context)
	CreateRoutinesByStationIds(ctx *gin.Context)
	UpdateMyRoutineById(ctx *gin.Context)
	UpdateMyRoutinesByIds(ctx *gin.Context)
	LinkRoutineTagById(ctx *gin.Context)
	LinkRoutineTagsByIds(ctx *gin.Context)
	LinkRoutineItemById(ctx *gin.Context)
	LinkRoutineItemsByIds(ctx *gin.Context)
	RestoreMyRoutineById(ctx *gin.Context)
	RestoreMyRoutinesByIds(ctx *gin.Context)
	DeleteMyRoutineById(ctx *gin.Context)
	DeleteMyRoutinesByIds(ctx *gin.Context)
	HardDeleteMyRoutineById(ctx *gin.Context)
	HardDeleteMyRoutinesByIds(ctx *gin.Context)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineStatusCount(ctx *gin.Context)
	VisualizeMyRoutinePeriodCount(ctx *gin.Context)
	VisualizeMyRoutineScheduledStartAtCount(ctx *gin.Context)
	VisualizeMyRoutineScheduledEndAtCount(ctx *gin.Context)

	/* ============================== GraphQL Methods ============================== */
	SearchRoutines(ctx *gin.Context)
}

type RoutineEndpoint struct {
	routineService routineservices.RoutineServiceInterface
}

func NewRoutineEndpoint(routineService routineservices.RoutineServiceInterface) RoutineEndpointInterface {
	return &RoutineEndpoint{routineService: routineService}
}

func (t *RoutineEndpoint) GetMyRoutineById(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRoutineByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.GetMyRoutineById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRoutineByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) GetMyRoutinesByStationId(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRoutinesByStationIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.GetMyRoutinesByStationId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRoutinesByStationIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) GetAllMyRoutinesByTimeRange(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetAllMyRoutinesByTimeRangeRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.GetAllMyRoutinesByTimeRange(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetAllMyRoutinesByTimeRangeResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) CreateRoutineByStationId(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRoutineByStationIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.CreateRoutineByStationId(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateRoutineByStationIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) CreateRoutinesByStationIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRoutinesByStationIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.CreateRoutinesByStationIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateRoutinesByStationIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) UpdateMyRoutineById(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRoutineByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.UpdateMyRoutineById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRoutineByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) UpdateMyRoutinesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRoutinesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.UpdateMyRoutinesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRoutinesByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) LinkRoutineTagById(ctx *gin.Context) {
	request := &cgateway.Request[capi.LinkRoutineTagByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.LinkRoutineTagById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.LinkRoutineTagByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) LinkRoutineTagsByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.LinkRoutineTagsByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.LinkRoutineTagsByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.LinkRoutineTagsByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) LinkRoutineItemById(ctx *gin.Context) {
	request := &cgateway.Request[capi.LinkRoutineItemByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.LinkRoutineItemById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.LinkRoutineItemByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) LinkRoutineItemsByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.LinkRoutineItemsByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.LinkRoutineItemsByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.LinkRoutineItemsByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) RestoreMyRoutineById(ctx *gin.Context) {
	request := &cgateway.Request[capi.RestoreMyRoutineByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.RestoreMyRoutineById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.RestoreMyRoutineByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) RestoreMyRoutinesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.RestoreMyRoutinesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.RestoreMyRoutinesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.RestoreMyRoutinesByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) DeleteMyRoutineById(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRoutineByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.DeleteMyRoutineById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRoutineByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) DeleteMyRoutinesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRoutinesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.DeleteMyRoutinesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRoutinesByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) HardDeleteMyRoutineById(ctx *gin.Context) {
	request := &cgateway.Request[capi.HardDeleteMyRoutineByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.HardDeleteMyRoutineById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.HardDeleteMyRoutineByIdResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) HardDeleteMyRoutinesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.HardDeleteMyRoutinesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.HardDeleteMyRoutinesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.HardDeleteMyRoutinesByIdsResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) VisualizeMyRoutineStatusCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineStatusCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.VisualizeMyRoutineStatusCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineStatusCountResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) VisualizeMyRoutinePeriodCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutinePeriodCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.VisualizeMyRoutinePeriodCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutinePeriodCountResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) VisualizeMyRoutineScheduledStartAtCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineScheduledStartAtCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.VisualizeMyRoutineScheduledStartAtCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineScheduledStartAtCountResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RoutineEndpoint) VisualizeMyRoutineScheduledEndAtCount(ctx *gin.Context) {
	request := &cgateway.Request[capi.VisualizeMyRoutineScheduledEndAtCountRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.routineService.VisualizeMyRoutineScheduledEndAtCount(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: struct{}{}, Exception: publicException})
		return
	}
	ctx.JSON(http.StatusOK, cgateway.Response[capi.VisualizeMyRoutineScheduledEndAtCountResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}
