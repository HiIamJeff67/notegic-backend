package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type RoutineControllerInterface interface {
	GetMyRoutineById(ctx *gin.Context, requestDto *capi.GetMyRoutineByIdRequestDto)
	GetMyRoutinesByStationId(ctx *gin.Context, requestDto *capi.GetMyRoutinesByStationIdRequestDto)
	GetAllMyRoutinesByTimeRange(ctx *gin.Context, requestDto *capi.GetAllMyRoutinesByTimeRangeRequestDto)
	CreateRoutineByStationId(ctx *gin.Context, requestDto *capi.CreateRoutineByStationIdRequestDto)
	CreateRoutinesByStationIds(ctx *gin.Context, requestDto *capi.CreateRoutinesByStationIdsRequestDto)
	UpdateMyRoutineById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineByIdRequestDto)
	UpdateMyRoutinesByIds(ctx *gin.Context, requestDto *capi.UpdateMyRoutinesByIdsRequestDto)
	LinkRoutineTagById(ctx *gin.Context, requestDto *capi.LinkRoutineTagByIdRequestDto)
	LinkRoutineTagsByIds(ctx *gin.Context, requestDto *capi.LinkRoutineTagsByIdsRequestDto)
	LinkRoutineItemById(ctx *gin.Context, requestDto *capi.LinkRoutineItemByIdRequestDto)
	LinkRoutineItemsByIds(ctx *gin.Context, requestDto *capi.LinkRoutineItemsByIdsRequestDto)
	RestoreMyRoutineById(ctx *gin.Context, requestDto *capi.RestoreMyRoutineByIdRequestDto)
	RestoreMyRoutinesByIds(ctx *gin.Context, requestDto *capi.RestoreMyRoutinesByIdsRequestDto)
	DeleteMyRoutineById(ctx *gin.Context, requestDto *capi.DeleteMyRoutineByIdRequestDto)
	DeleteMyRoutinesByIds(ctx *gin.Context, requestDto *capi.DeleteMyRoutinesByIdsRequestDto)
	HardDeleteMyRoutineById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineByIdRequestDto)
	HardDeleteMyRoutinesByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutinesByIdsRequestDto)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineStatusCountRequestDto)
	VisualizeMyRoutinePeriodCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutinePeriodCountRequestDto)
	VisualizeMyRoutineScheduledStartAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineScheduledStartAtCountRequestDto)
	VisualizeMyRoutineScheduledEndAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineScheduledEndAtCountRequestDto)
}

type RoutineController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRoutineController(coreAdapter *coreadapters.CoreAdapter) RoutineControllerInterface {
	return &RoutineController{coreAdapter: coreAdapter}
}

func (c *RoutineController) GetMyRoutineById(ctx *gin.Context, requestDto *capi.GetMyRoutineByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetMyRoutineByIdRequestDto, capi.GetMyRoutineByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyRoutineByIdOperation,
		"/core/v1/routines/get-by-id",
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

func (c *RoutineController) GetMyRoutinesByStationId(ctx *gin.Context, requestDto *capi.GetMyRoutinesByStationIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetMyRoutinesByStationIdRequestDto, capi.GetMyRoutinesByStationIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyRoutinesByStationIdOperation,
		"/core/v1/routines/get-by-station-id",
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

func (c *RoutineController) GetAllMyRoutinesByTimeRange(ctx *gin.Context, requestDto *capi.GetAllMyRoutinesByTimeRangeRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetAllMyRoutinesByTimeRangeRequestDto, capi.GetAllMyRoutinesByTimeRangeResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetAllMyRoutinesByTimeRangeOperation,
		"/core/v1/routines/get-all-by-time-range",
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

func (c *RoutineController) CreateRoutineByStationId(ctx *gin.Context, requestDto *capi.CreateRoutineByStationIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.CreateRoutineByStationIdRequestDto, capi.CreateRoutineByStationIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutineByStationIdOperation,
		"/core/v1/routines/create-by-station-id",
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

func (c *RoutineController) CreateRoutinesByStationIds(ctx *gin.Context, requestDto *capi.CreateRoutinesByStationIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.CreateRoutinesByStationIdsRequestDto, capi.CreateRoutinesByStationIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateRoutinesByStationIdsOperation,
		"/core/v1/routines/create-many-by-station-ids",
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

func (c *RoutineController) UpdateMyRoutineById(ctx *gin.Context, requestDto *capi.UpdateMyRoutineByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.UpdateMyRoutineByIdRequestDto, capi.UpdateMyRoutineByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRoutineByIdOperation,
		"/core/v1/routines/update",
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

func (c *RoutineController) UpdateMyRoutinesByIds(ctx *gin.Context, requestDto *capi.UpdateMyRoutinesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.UpdateMyRoutinesByIdsRequestDto, capi.UpdateMyRoutinesByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRoutinesByIdsOperation,
		"/core/v1/routines/update-many",
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

func (c *RoutineController) LinkRoutineTagById(ctx *gin.Context, requestDto *capi.LinkRoutineTagByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.LinkRoutineTagByIdRequestDto, capi.LinkRoutineTagByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LinkRoutineTagByIdOperation,
		"/core/v1/routines/link-tag",
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

func (c *RoutineController) LinkRoutineTagsByIds(ctx *gin.Context, requestDto *capi.LinkRoutineTagsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.LinkRoutineTagsByIdsRequestDto, capi.LinkRoutineTagsByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LinkRoutineTagsByIdsOperation,
		"/core/v1/routines/link-tags",
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

func (c *RoutineController) LinkRoutineItemById(ctx *gin.Context, requestDto *capi.LinkRoutineItemByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.LinkRoutineItemByIdRequestDto, capi.LinkRoutineItemByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LinkRoutineItemByIdOperation,
		"/core/v1/routines/link-item",
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

func (c *RoutineController) LinkRoutineItemsByIds(ctx *gin.Context, requestDto *capi.LinkRoutineItemsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.LinkRoutineItemsByIdsRequestDto, capi.LinkRoutineItemsByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LinkRoutineItemsByIdsOperation,
		"/core/v1/routines/link-items",
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

func (c *RoutineController) RestoreMyRoutineById(ctx *gin.Context, requestDto *capi.RestoreMyRoutineByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.RestoreMyRoutineByIdRequestDto, capi.RestoreMyRoutineByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyRoutineByIdOperation,
		"/core/v1/routines/restore",
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

func (c *RoutineController) RestoreMyRoutinesByIds(ctx *gin.Context, requestDto *capi.RestoreMyRoutinesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.RestoreMyRoutinesByIdsRequestDto, capi.RestoreMyRoutinesByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyRoutinesByIdsOperation,
		"/core/v1/routines/restore-many",
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

func (c *RoutineController) DeleteMyRoutineById(ctx *gin.Context, requestDto *capi.DeleteMyRoutineByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.DeleteMyRoutineByIdRequestDto, capi.DeleteMyRoutineByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRoutineByIdOperation,
		"/core/v1/routines/delete",
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

func (c *RoutineController) DeleteMyRoutinesByIds(ctx *gin.Context, requestDto *capi.DeleteMyRoutinesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.DeleteMyRoutinesByIdsRequestDto, capi.DeleteMyRoutinesByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRoutinesByIdsOperation,
		"/core/v1/routines/delete-many",
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

func (c *RoutineController) HardDeleteMyRoutineById(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutineByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.HardDeleteMyRoutineByIdRequestDto, capi.HardDeleteMyRoutineByIdResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutineByIdOperation,
		"/core/v1/routines/hard-delete",
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

func (c *RoutineController) HardDeleteMyRoutinesByIds(ctx *gin.Context, requestDto *capi.HardDeleteMyRoutinesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.HardDeleteMyRoutinesByIdsRequestDto, capi.HardDeleteMyRoutinesByIdsResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.HardDeleteMyRoutinesByIdsOperation,
		"/core/v1/routines/hard-delete-many",
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

func (c *RoutineController) VisualizeMyRoutineStatusCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineStatusCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineStatusCountRequestDto, capi.VisualizeMyRoutineStatusCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineStatusCountOperation,
		"/core/v1/routines/visualize-status-count",
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

func (c *RoutineController) VisualizeMyRoutinePeriodCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutinePeriodCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutinePeriodCountRequestDto, capi.VisualizeMyRoutinePeriodCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutinePeriodCountOperation,
		"/core/v1/routines/visualize-period-count",
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

func (c *RoutineController) VisualizeMyRoutineScheduledStartAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineScheduledStartAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineScheduledStartAtCountRequestDto, capi.VisualizeMyRoutineScheduledStartAtCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineScheduledStartAtCountOperation,
		"/core/v1/routines/visualize-scheduled-start-at-count",
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

func (c *RoutineController) VisualizeMyRoutineScheduledEndAtCount(ctx *gin.Context, requestDto *capi.VisualizeMyRoutineScheduledEndAtCountRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.VisualizeMyRoutineScheduledEndAtCountRequestDto, capi.VisualizeMyRoutineScheduledEndAtCountResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.VisualizeMyRoutineScheduledEndAtCountOperation,
		"/core/v1/routines/visualize-scheduled-end-at-count",
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
