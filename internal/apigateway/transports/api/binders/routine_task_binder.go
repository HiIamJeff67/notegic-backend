package binders

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/controllers"
)

type RoutineTaskBinderInterface interface {
	BindGetMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.GetMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutineTasksByRoutineIds(controllerFunc controllers.Func[*apicontract.GetAllMyRoutineTasksByRoutineIdsRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutineTasks(controllerFunc controllers.Func[*apicontract.GetAllMyRoutineTasksRequestDto]) gin.HandlerFunc
	BindCreateRoutineTaskByRoutineId(controllerFunc controllers.Func[*apicontract.CreateRoutineTaskByRoutineIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.UpdateMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindPauseMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.PauseMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindResumeMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.ResumeMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTasksByIds(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineTasksByIdsRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyRoutineTaskStatusCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskStatusCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskPurposeCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskPurposeCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskScheduledAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskScheduledAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskActualStartedAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskActualStartedAtCountRequestDto]) gin.HandlerFunc
	BindVisualizeMyRoutineTaskActualEndedAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskActualEndedAtCountRequestDto]) gin.HandlerFunc
}

type RoutineTaskBinder struct{}

func NewRoutineTaskBinder() RoutineTaskBinderInterface { return &RoutineTaskBinder{} }

func bindRoutineTaskJSON[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidDto("RoutineTask").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}

func parseRoutineTaskUUID(ctx *gin.Context, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
		return uuid.Nil, false
	}
	return value, true
}
func parseRoutineTaskBool(ctx *gin.Context, name string) (*bool, bool) {
	valueString := ctx.Query(name)
	if valueString == "" {
		return nil, true
	}
	value, err := strconv.ParseBool(valueString)
	if err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
		return nil, false
	}
	return &value, true
}
func parseRoutineTaskPermission(ctx *gin.Context) (enumcontract.AccessControlPermission, bool) {
	permission := enumcontract.AccessControlPermission(ctx.Query("permission"))
	switch permission {
	case enumcontract.AccessControlPermission_Read,
		enumcontract.AccessControlPermission_Write,
		enumcontract.AccessControlPermission_Admin,
		enumcontract.AccessControlPermission_Owner:
		return permission, true
	default:
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("RoutineTask"), ctx)
		return "", false
	}
}
func parseRoutineTaskTime(ctx *gin.Context, name string) (time.Time, bool) {
	value, err := time.Parse(time.RFC3339, ctx.Query(name))
	if err != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
		return time.Time{}, false
	}
	return value, true
}

func (b *RoutineTaskBinder) BindGetMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.GetMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetMyRoutineTaskByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		isDeleted, ok := parseRoutineTaskBool(ctx, "isDeleted")
		if !ok {
			return
		}
		requestDto.Param.IsDeleted = isDeleted
		value, ok := parseRoutineTaskUUID(ctx, "routine-task-id")
		if !ok {
			return
		}
		requestDto.Param.RoutineTaskId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindGetAllMyRoutineTasksByRoutineIds(controllerFunc controllers.Func[*apicontract.GetAllMyRoutineTasksByRoutineIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetAllMyRoutineTasksByRoutineIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		areDeleted, ok := parseRoutineTaskBool(ctx, "areDeleted")
		if !ok {
			return
		}
		requestDto.Param.AreDeleted = areDeleted
		if ctx.Query("areDeleted") != "" && requestDto.Param.AreDeleted == nil {
			return
		}
		routineIdValues := ctx.QueryArray("routineIds")
		if len(routineIdValues) == 1 {
			routineIdValues = strings.Split(routineIdValues[0], ",")
		}
		requestDto.Param.RoutineIds = make([]uuid.UUID, len(routineIdValues))
		for index, value := range routineIdValues {
			parsed, err := uuid.Parse(value)
			if err != nil {
				exceptionwriter.SafelyAbortAndResponseWithJSON(exceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
				return
			}
			requestDto.Param.RoutineIds[index] = parsed
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindGetAllMyRoutineTasks(controllerFunc controllers.Func[*apicontract.GetAllMyRoutineTasksRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.GetAllMyRoutineTasksRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		areDeleted, ok := parseRoutineTaskBool(ctx, "areDeleted")
		if !ok {
			return
		}
		requestDto.Param.AreDeleted = areDeleted
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindCreateRoutineTaskByRoutineId(controllerFunc controllers.Func[*apicontract.CreateRoutineTaskByRoutineIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.CreateRoutineTaskByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineTaskUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineId = value
		bindRoutineTaskJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineTaskBinder) BindUpdateMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.UpdateMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.UpdateMyRoutineTaskByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineTaskUUID(ctx, "routine-task-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineTaskId = value
		bindRoutineTaskJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineTaskBinder) BindPauseMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.PauseMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.PauseMyRoutineTaskByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineTaskUUID(ctx, "routine-task-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineTaskId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindResumeMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.ResumeMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.ResumeMyRoutineTaskByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineTaskUUID(ctx, "routine-task-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineTaskId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindHardDeleteMyRoutineTaskById(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.HardDeleteMyRoutineTaskByIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		value, ok := parseRoutineTaskUUID(ctx, "routine-task-id")
		if !ok {
			return
		}
		requestDto.Body.RoutineTaskId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindHardDeleteMyRoutineTasksByIds(controllerFunc controllers.Func[*apicontract.HardDeleteMyRoutineTasksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.HardDeleteMyRoutineTasksByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineTaskJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskStatusCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskStatusCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineTaskStatusCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutineTaskPermission(ctx)
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskPurposeCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskPurposeCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineTaskPurposeCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutineTaskPermission(ctx)
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskScheduledAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskScheduledAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineTaskScheduledAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutineTaskPermission(ctx)
		if !ok {
			return
		}
		requestDto.Param.TimeHourUnit, _ = strconv.Atoi(ctx.Query("timeHourUnit"))
		requestDto.Param.QueryRangeStartedAt, ok = parseRoutineTaskTime(ctx, "queryRangeStartedAt")
		if !ok {
			return
		}
		requestDto.Param.QueryRangeEndedAt, ok = parseRoutineTaskTime(ctx, "queryRangeEndedAt")
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskActualStartedAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskActualStartedAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineTaskActualStartedAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutineTaskPermission(ctx)
		if !ok {
			return
		}
		requestDto.Param.TimeHourUnit, _ = strconv.Atoi(ctx.Query("timeHourUnit"))
		requestDto.Param.QueryRangeStartedAt, ok = parseRoutineTaskTime(ctx, "queryRangeStartedAt")
		if !ok {
			return
		}
		requestDto.Param.QueryRangeEndedAt, ok = parseRoutineTaskTime(ctx, "queryRangeEndedAt")
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskActualEndedAtCount(controllerFunc controllers.Func[*apicontract.VisualizeMyRoutineTaskActualEndedAtCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &apicontract.VisualizeMyRoutineTaskActualEndedAtCountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		ok := true
		requestDto.Param.Permission, ok = parseRoutineTaskPermission(ctx)
		if !ok {
			return
		}
		requestDto.Param.TimeHourUnit, _ = strconv.Atoi(ctx.Query("timeHourUnit"))
		requestDto.Param.QueryRangeStartedAt, ok = parseRoutineTaskTime(ctx, "queryRangeStartedAt")
		if !ok {
			return
		}
		requestDto.Param.QueryRangeEndedAt, ok = parseRoutineTaskTime(ctx, "queryRangeEndedAt")
		if !ok {
			return
		}
		controllerFunc(ctx, requestDto)
		return
	}
}
