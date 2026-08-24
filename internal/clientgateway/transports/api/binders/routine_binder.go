package binders

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type RoutineBinderInterface interface {
	BindGetMyRoutineById(controllerFunc controllers.Func[*apicontract.GetMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindGetMyRoutinesByStationId(controllerFunc controllers.Func[*apicontract.GetMyRoutinesByStationIdRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutinesByTimeRange(controllerFunc controllers.Func[*apicontract.GetAllMyRoutinesByTimeRangeRequestDto]) gin.HandlerFunc
	BindCreateRoutineByStationId(controllerFunc controllers.Func[*apicontract.CreateRoutineByStationIdRequestDto]) gin.HandlerFunc
	BindCreateRoutinesByStationIds(controllerFunc controllers.Func[*apicontract.CreateRoutinesByStationIdsRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineById(controllerFunc controllers.Func[*apicontract.UpdateMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.UpdateMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindLinkRoutineTagById(controllerFunc controllers.Func[*apicontract.LinkRoutineTagByIdRequestDto]) gin.HandlerFunc
	BindLinkRoutineTagsByIds(controllerFunc controllers.Func[*apicontract.LinkRoutineTagsByIdsRequestDto]) gin.HandlerFunc
	BindLinkRoutineItemById(controllerFunc controllers.Func[*apicontract.LinkRoutineItemByIdRequestDto]) gin.HandlerFunc
	BindLinkRoutineItemsByIds(controllerFunc controllers.Func[*apicontract.LinkRoutineItemsByIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyRoutineById(controllerFunc controllers.Func[*apicontract.RestoreMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.RestoreMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyRoutineById(controllerFunc controllers.Func[*apicontract.DeleteMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.DeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineById(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyRoutineStatusCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineStatusCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutinePeriodCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutinePeriodCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineScheduledStartAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineScheduledStartAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineScheduledEndAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineScheduledEndAtCountRequestDto]) gin.HandlerFunc
}

type RoutineBinder struct{}

func NewRoutineBinder() RoutineBinderInterface { return &RoutineBinder{} }

func bindRoutineJSON[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidDto("Routine").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}

func parseRoutineUUID(ctx *gin.Context, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("Routine").WithOrigin(err), ctx)
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
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("Routine").WithOrigin(err), ctx)
		return nil, false
	}
	return &value, true
}

func parseRoutinePermission(ctx *gin.Context) (enumcontract.AccessControlPermission, bool) {
	permission := enumcontract.AccessControlPermission(ctx.Query("permission"))
	switch permission {
	case enumcontract.AccessControlPermission_Read,
		enumcontract.AccessControlPermission_Write,
		enumcontract.AccessControlPermission_Admin,
		enumcontract.AccessControlPermission_Owner:
		return permission, true
	default:
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("Routine"), ctx)
		return "", false
	}
}

func parseRoutineTime(ctx *gin.Context, name string) (time.Time, bool) {
	value, err := time.Parse(time.RFC3339, ctx.Query(name))
	if err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("Routine").WithOrigin(err), ctx)
		return time.Time{}, false
	}
	return value, true
}

func (b *RoutineBinder) BindGetMyRoutineById(controllerFunc controllers.Func[*apicontract.GetMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetMyRoutineByIdRequestDto{}
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

func (b *RoutineBinder) BindGetMyRoutinesByStationId(controllerFunc controllers.Func[*apicontract.GetMyRoutinesByStationIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetMyRoutinesByStationIdRequestDto{}
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

func (b *RoutineBinder) BindGetAllMyRoutinesByTimeRange(controllerFunc controllers.Func[*apicontract.GetAllMyRoutinesByTimeRangeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetAllMyRoutinesByTimeRangeRequestDto{}
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
				exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("Routine").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.StationIds[index] = parsed
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindCreateRoutineByStationId(controllerFunc controllers.Func[*apicontract.CreateRoutineByStationIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.CreateRoutineByStationIdRequestDto{}
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

func (b *RoutineBinder) BindCreateRoutinesByStationIds(controllerFunc controllers.Func[*apicontract.CreateRoutinesByStationIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.CreateRoutinesByStationIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindUpdateMyRoutineById(controllerFunc controllers.Func[*apicontract.UpdateMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.UpdateMyRoutineByIdRequestDto{}
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

func (b *RoutineBinder) BindUpdateMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.UpdateMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.UpdateMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineTagById(controllerFunc controllers.Func[*apicontract.LinkRoutineTagByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.LinkRoutineTagByIdRequestDto{}
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

func (b *RoutineBinder) BindLinkRoutineTagsByIds(controllerFunc controllers.Func[*apicontract.LinkRoutineTagsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.LinkRoutineTagsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineItemById(controllerFunc controllers.Func[*apicontract.LinkRoutineItemByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.LinkRoutineItemByIdRequestDto{}
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
		requestDto.Body.ItemType = enumcontract.ItemType(ctx.Query("itemType"))
		requestDto.Body.IsUnlink = ctx.Query("isUnlink") == "true"
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineBinder) BindLinkRoutineItemsByIds(controllerFunc controllers.Func[*apicontract.LinkRoutineItemsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.LinkRoutineItemsByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindRestoreMyRoutineById(controllerFunc controllers.Func[*apicontract.RestoreMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.RestoreMyRoutineByIdRequestDto{}
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

func (b *RoutineBinder) BindRestoreMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.RestoreMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.RestoreMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindDeleteMyRoutineById(controllerFunc controllers.Func[*apicontract.DeleteMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.DeleteMyRoutineByIdRequestDto{}
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

func (b *RoutineBinder) BindDeleteMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.DeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.DeleteMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindHardDeleteMyRoutineById(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.HardDeleteMyRoutineByIdRequestDto{}
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

func (b *RoutineBinder) BindHardDeleteMyRoutinesByIds(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutinesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.HardDeleteMyRoutinesByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineBinder) BindVisualizeMyRoutineStatusCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineStatusCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineStatusCountRequestDto{}
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

func (b *RoutineBinder) BindVisualizeMyRoutinePeriodCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutinePeriodCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutinePeriodCountRequestDto{}
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

func (b *RoutineBinder) BindVisualizeMyRoutineScheduledStartAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineScheduledStartAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineScheduledStartAtCountRequestDto{}
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

func (b *RoutineBinder) BindVisualizeMyRoutineScheduledEndAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineScheduledEndAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineScheduledEndAtCountRequestDto{}
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
