package binders

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-records"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type RoutineTaskRecordBinderInterface interface {
	BindGetAllMyRoutineTaskRecordsByRoutineTaskId(controllerFunc controllers.Func[*capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyRoutineTaskRecordStatusCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskRecordPurposeCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskRecordScheduledAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskRecordActualStartedAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskRecordActualEndedAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto]) gin.HandlerFunc
}

type RoutineTaskRecordBinder struct{}

func NewRoutineTaskRecordBinder() RoutineTaskRecordBinderInterface {
	return &RoutineTaskRecordBinder{}
}

/* ============================== Auxiliary Functions ============================== */

type routineTaskRecordVisualizationParams struct {
	permission     string
	routineTaskIds []uuid.UUID
}

func parseRoutineTaskRecordVisualizationParams(ctx *gin.Context) (*routineTaskRecordVisualizationParams, *cexceptions.Exception) {
	permission := ctx.Query("permission")
	if permission == "" {
		return nil, cexceptions.InvalidInput("RoutineTask").WithOrigin(fmt.Errorf("permission is required"))
	}

	params := &routineTaskRecordVisualizationParams{
		permission: permission,
	}
	for _, routineTaskIdsValue := range ctx.QueryArray("routineTaskIds") {
		for _, routineTaskIdValue := range strings.Split(routineTaskIdsValue, ",") {
			trimmedRoutineTaskIdValue := strings.TrimSpace(routineTaskIdValue)
			if trimmedRoutineTaskIdValue == "" {
				continue
			}
			routineTaskId, err := uuid.Parse(trimmedRoutineTaskIdValue)
			if err != nil {
				return nil, cexceptions.InvalidInput("RoutineTask").WithOrigin(err)
			}
			params.routineTaskIds = append(params.routineTaskIds, routineTaskId)
		}
	}

	return params, nil
}

func parseRoutineTaskRecordVisualizationTimeRange(ctx *gin.Context) (int, time.Time, time.Time, *cexceptions.Exception) {
	timeHourUnitString := ctx.Query("timeHourUnit")
	if timeHourUnitString == "" {
		return 0, time.Time{}, time.Time{}, cexceptions.InvalidInput("RoutineTask").WithOrigin(fmt.Errorf("timeHourUnit is required"))
	}
	timeHourUnit, err := strconv.Atoi(timeHourUnitString)
	if err != nil {
		return 0, time.Time{}, time.Time{}, cexceptions.InvalidInput("RoutineTask").WithOrigin(err)
	}

	queryRangeStartedAt, err := time.Parse(time.RFC3339, ctx.Query("queryRangeStartedAt"))
	if err != nil {
		return 0, time.Time{}, time.Time{}, cexceptions.InvalidInput("RoutineTask").WithOrigin(fmt.Errorf("queryRangeStartedAt must be an RFC3339 timestamp: %w", err))
	}
	queryRangeEndedAt, err := time.Parse(time.RFC3339, ctx.Query("queryRangeEndedAt"))
	if err != nil {
		return 0, time.Time{}, time.Time{}, cexceptions.InvalidInput("RoutineTask").WithOrigin(fmt.Errorf("queryRangeEndedAt must be an RFC3339 timestamp: %w", err))
	}

	return timeHourUnit, queryRangeStartedAt, queryRangeEndedAt, nil
}

/* ============================== Service Methods for RoutineTaskRecord ============================== */

func (b *RoutineTaskRecordBinder) BindGetAllMyRoutineTaskRecordsByRoutineTaskId(controllerFunc controllers.Func[*capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetAllMyRoutineTaskRecordsByRoutineTaskIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		routineTaskId, err := uuid.Parse(ctx.Param("routine-task-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RoutineTaskId = routineTaskId

		limitString := ctx.Query("limit")
		if limitString != "" {
			limit, err := strconv.Atoi(limitString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.Limit = limit
		}

		controllerFunc(ctx, requestDto)
	}
}

/* ============================== Visualization Methods ============================== */

func (b *RoutineTaskRecordBinder) BindVisualizeMyRoutineTaskRecordStatusCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		params, exception := parseRoutineTaskRecordVisualizationParams(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		requestDto.Param.Permission = params.permission
		requestDto.Param.RoutineTaskIds = params.routineTaskIds
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskRecordBinder) BindVisualizeMyRoutineTaskRecordPurposeCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		params, exception := parseRoutineTaskRecordVisualizationParams(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		requestDto.Param.Permission = params.permission
		requestDto.Param.RoutineTaskIds = params.routineTaskIds
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskRecordBinder) BindVisualizeMyRoutineTaskRecordScheduledAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		params, exception := parseRoutineTaskRecordVisualizationParams(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		timeHourUnit, queryRangeStartedAt, queryRangeEndedAt, exception := parseRoutineTaskRecordVisualizationTimeRange(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		requestDto.Param.Permission = params.permission
		requestDto.Param.RoutineTaskIds = params.routineTaskIds
		requestDto.Param.TimeHourUnit = timeHourUnit
		requestDto.Param.QueryRangeStartedAt = queryRangeStartedAt
		requestDto.Param.QueryRangeEndedAt = queryRangeEndedAt
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskRecordBinder) BindVisualizeMyRoutineTaskRecordActualStartedAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		params, exception := parseRoutineTaskRecordVisualizationParams(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		timeHourUnit, queryRangeStartedAt, queryRangeEndedAt, exception := parseRoutineTaskRecordVisualizationTimeRange(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		requestDto.Param.Permission = params.permission
		requestDto.Param.RoutineTaskIds = params.routineTaskIds
		requestDto.Param.TimeHourUnit = timeHourUnit
		requestDto.Param.QueryRangeStartedAt = queryRangeStartedAt
		requestDto.Param.QueryRangeEndedAt = queryRangeEndedAt
		controllerFunc(ctx, requestDto)
	}
}

func (b *RoutineTaskRecordBinder) BindVisualizeMyRoutineTaskRecordActualEndedAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		params, exception := parseRoutineTaskRecordVisualizationParams(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		timeHourUnit, queryRangeStartedAt, queryRangeEndedAt, exception := parseRoutineTaskRecordVisualizationTimeRange(ctx)
		if exception != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}
		requestDto.Param.Permission = params.permission
		requestDto.Param.RoutineTaskIds = params.routineTaskIds
		requestDto.Param.TimeHourUnit = timeHourUnit
		requestDto.Param.QueryRangeStartedAt = queryRangeStartedAt
		requestDto.Param.QueryRangeEndedAt = queryRangeEndedAt
		controllerFunc(ctx, requestDto)
	}
}
