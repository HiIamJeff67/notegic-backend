package binders

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/controllers"
)

type RoutineBinderInterface interface {
	BindGetMyRoutineById(controllerFunc controllers.Func[*capi.GetMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindGetMyRoutinesByStationId(controllerFunc controllers.Func[*capi.GetMyRoutinesByStationIdRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutinesByTimeRange(controllerFunc controllers.Func[*capi.GetAllMyRoutinesByTimeRangeRequestDto]) gin.HandlerFunc
	BindCreateRoutineByStationId(controllerFunc controllers.Func[*capi.CreateRoutineByStationIdRequestDto]) gin.HandlerFunc
	BindCreateRoutinesByStationIds(controllerFunc controllers.Func[*capi.CreateRoutinesByStationIdsRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineById(controllerFunc controllers.Func[*capi.UpdateMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutinesByIds(controllerFunc controllers.Func[*capi.UpdateMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindLinkRoutineTagById(controllerFunc controllers.Func[*capi.LinkRoutineTagByIdRequestDto]) gin.HandlerFunc
	BindLinkRoutineTagsByIds(controllerFunc controllers.Func[*capi.LinkRoutineTagsByIdsRequestDto]) gin.HandlerFunc
	BindLinkRoutineItemById(controllerFunc controllers.Func[*capi.LinkRoutineItemByIdRequestDto]) gin.HandlerFunc
	BindLinkRoutineItemsByIds(controllerFunc controllers.Func[*capi.LinkRoutineItemsByIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyRoutineById(controllerFunc controllers.Func[*capi.RestoreMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyRoutinesByIds(controllerFunc controllers.Func[*capi.RestoreMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyRoutineById(controllerFunc controllers.Func[*capi.DeleteMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyRoutinesByIds(controllerFunc controllers.Func[*capi.DeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutinesByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyRoutineStatusCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineStatusCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutinePeriodCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutinePeriodCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineScheduledStartAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineScheduledStartAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineScheduledEndAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineScheduledEndAtCountRequestDto]) gin.HandlerFunc
}

type RoutineBinder struct{}

func NewRoutineBinder() RoutineBinderInterface { return &RoutineBinder{} }

func bindRoutineJSON[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Routine").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}

func parseRoutineUUID(ctx *gin.Context, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Routine").WithOrigin(err), ctx)
		return uuid.Nil, false
	}
	return value, true
}

func parseRoutineBool(ctx *gin.Context, name string) (*bool, bool) {
	valueString := ctx.Query(name)
	if valueString == "" {
		return nil, true
	}
	value, err := strconv.ParseBool(valueString)
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Routine").WithOrigin(err), ctx)
		return nil, false
	}
	return &value, true
}

func parseRoutinePermission(ctx *gin.Context) (cenums.AccessControlPermission, bool) {
	permission := cenums.AccessControlPermission(ctx.Query("permission"))
	switch permission {
	case cenums.AccessControlPermission_Read,
		cenums.AccessControlPermission_Write,
		cenums.AccessControlPermission_Admin,
		cenums.AccessControlPermission_Owner:
		return permission, true
	default:
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Routine"), ctx)
		return "", false
	}
}

func parseRoutineTime(ctx *gin.Context, name string) (time.Time, bool) {
	value, err := time.Parse(time.RFC3339, ctx.Query(name))
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Routine").WithOrigin(err), ctx)
		return time.Time{}, false
	}
	return value, true
}

func (b *RoutineBinder) BindGetMyRoutineById(controllerFunc controllers.Func[*capi.GetMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRoutineByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		isDeleted, ok := parseRoutineBool(ctx, "isDeleted")
		if !ok {
			return
		}
		requestDto.Param.IsDeleted = isDeleted
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Param.RoutineId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindGetMyRoutinesByStationId(controllerFunc controllers.Func[*capi.GetMyRoutinesByStationIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRoutinesByStationIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		areDeleted, ok := parseRoutineBool(ctx, "areDeleted")
		if !ok {
			return
		}
		requestDto.Param.AreDeleted = areDeleted
		value, ok := parseRoutineUUID(ctx, "station-id")
		if !ok {
			return
		}
		requestDto.Param.StationId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindGetAllMyRoutinesByTimeRange(controllerFunc controllers.Func[*capi.GetAllMyRoutinesByTimeRangeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetAllMyRoutinesByTimeRangeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.AreDeleted, ok = parseRoutineBool(ctx, "areDeleted")
		if requestDto.Param.AreDeleted == nil && ctx.Query("areDeleted") != "" {
			return
		}
		requestDto.Param.From, ok = parseRoutineTime(ctx, "from")
		if !ok {
			return
		}
		requestDto.Param.To, ok = parseRoutineTime(ctx, "to")
		if !ok {
			return
		}
		stationIdValues := ctx.QueryArray("stationIds")
		if len(stationIdValues) == 1 {
			stationIdValues = strings.Split(stationIdValues[0], ",")
		}
		requestDto.Param.StationIds = make([]uuid.UUID, len(stationIdValues))
		for index, value := range stationIdValues {
			parsed, err := uuid.Parse(value)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Routine").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.StationIds[index] = parsed
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindCreateRoutineByStationId(controllerFunc controllers.Func[*capi.CreateRoutineByStationIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineByStationIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "station-id")
		if !ok {
			return
		}
		requestDto.Body.StationId = value
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindCreateRoutinesByStationIds(controllerFunc controllers.Func[*capi.CreateRoutinesByStationIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutinesByStationIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindUpdateMyRoutineById(controllerFunc controllers.Func[*capi.UpdateMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRoutineByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindUpdateMyRoutinesByIds(controllerFunc controllers.Func[*capi.UpdateMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineTagById(controllerFunc controllers.Func[*capi.LinkRoutineTagByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LinkRoutineTagByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		value, ok = parseRoutineUUID(ctx, "routine-tag-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineTagId = value
		requestDto.Body.IsUnlink = ctx.Query("isUnlink") == "true"
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineTagsByIds(controllerFunc controllers.Func[*capi.LinkRoutineTagsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LinkRoutineTagsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineItemById(controllerFunc controllers.Func[*capi.LinkRoutineItemByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LinkRoutineItemByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		value, ok = parseRoutineUUID(ctx, "item-id")
		if !ok {
			return
		}
		requestDto.Body.ItemId = value
		requestDto.Body.ItemType = cenums.ItemType(ctx.Query("itemType"))
		requestDto.Body.IsUnlink = ctx.Query("isUnlink") == "true"
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineItemsByIds(controllerFunc controllers.Func[*capi.LinkRoutineItemsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LinkRoutineItemsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindRestoreMyRoutineById(controllerFunc controllers.Func[*capi.RestoreMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.RestoreMyRoutineByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindRestoreMyRoutinesByIds(controllerFunc controllers.Func[*capi.RestoreMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.RestoreMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindDeleteMyRoutineById(controllerFunc controllers.Func[*capi.DeleteMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteMyRoutineByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindDeleteMyRoutinesByIds(controllerFunc controllers.Func[*capi.DeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindHardDeleteMyRoutineById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutineByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindHardDeleteMyRoutinesByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindVisualizeMyRoutineStatusCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineStatusCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineStatusCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutinePermission(ctx)
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindVisualizeMyRoutinePeriodCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutinePeriodCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutinePeriodCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutinePermission(ctx)
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindVisualizeMyRoutineScheduledStartAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineScheduledStartAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineScheduledStartAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutinePermission(ctx)
		if !ok {
			return
		}
		requestDto.Param.TimeHourUnit, _ = strconv.Atoi(ctx.Query("timeHourUnit"))
		requestDto.Param.QueryRangeStartedAt, ok = parseRoutineTime(ctx, "queryRangeStartedAt")
		if !ok {
			return
		}
		requestDto.Param.QueryRangeEndedAt, ok = parseRoutineTime(ctx, "queryRangeEndedAt")
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindVisualizeMyRoutineScheduledEndAtCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineScheduledEndAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineScheduledEndAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutinePermission(ctx)
		if !ok {
			return
		}
		requestDto.Param.TimeHourUnit, _ = strconv.Atoi(ctx.Query("timeHourUnit"))
		requestDto.Param.QueryRangeStartedAt, ok = parseRoutineTime(ctx, "queryRangeStartedAt")
		if !ok {
			return
		}
		requestDto.Param.QueryRangeEndedAt, ok = parseRoutineTime(ctx, "queryRangeEndedAt")
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}
