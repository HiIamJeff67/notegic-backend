package binders

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type RoutineTaskBinderInterface interface {
	BindGetMyRoutineTaskById(controllerFunc controllers.Func[*capi.GetMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindGetMyRoutineTasksByRoutineId(controllerFunc controllers.Func[*capi.GetMyRoutineTasksByRoutineIdRequestDto]) gin.HandlerFunc
	BindGetAllMyRoutineTasks(controllerFunc controllers.Func[*capi.GetAllMyRoutineTasksRequestDto]) gin.HandlerFunc
	BindCreateRoutineTaskByRoutineId(controllerFunc controllers.Func[*capi.CreateRoutineTaskByRoutineIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRoutineTaskById(controllerFunc controllers.Func[*capi.UpdateMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTaskById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTaskByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyRoutineTasksByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTasksByIdsRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyRoutineTaskPurposeCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskPurposeCountRequestDto]) gin.HandlerFunc
}

type RoutineTaskBinder struct{}

func NewRoutineTaskBinder() RoutineTaskBinderInterface { return &RoutineTaskBinder{} }

func bindRoutineTaskJSON[T any](ctx *gin.Context, requestDto *T, body any, controllerFunc controllers.Func[*T]) {
	if err := ctx.ShouldBindJSON(body); err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("RoutineTask").WithOrigin(err), ctx)
		return
	}
	controllerFunc(ctx, requestDto)
}

func parseRoutineTaskUUID(ctx *gin.Context, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
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
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTask").WithOrigin(err), ctx)
		return nil, false
	}
	return &value, true
}
func parseRoutineTaskPermission(ctx *gin.Context) (cenums.AccessControlPermission, bool) {
	permission := cenums.AccessControlPermission(ctx.Query("permission"))
	switch permission {
	case cenums.AccessControlPermission_Read,
		cenums.AccessControlPermission_Write,
		cenums.AccessControlPermission_Admin,
		cenums.AccessControlPermission_Owner:
		return permission, true
	default:
		sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("RoutineTask"), ctx)
		return "", false
	}
}
func (b *RoutineTaskBinder) BindGetMyRoutineTaskById(controllerFunc controllers.Func[*capi.GetMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRoutineTaskByIdRequestDto{}
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

func (b *RoutineTaskBinder) BindGetMyRoutineTasksByRoutineId(controllerFunc controllers.Func[*capi.GetMyRoutineTasksByRoutineIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRoutineTasksByRoutineIdRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		areDeleted, ok := parseRoutineTaskBool(ctx, "areDeleted")
		if !ok {
			return
		}
		requestDto.Param.AreDeleted = areDeleted
		value, ok := parseRoutineTaskUUID(ctx, "routine-id")
		if !ok {
			return
		}
		requestDto.Param.RoutineId = value
		controllerFunc(ctx, requestDto)
		return
	}
}

func (b *RoutineTaskBinder) BindGetAllMyRoutineTasks(controllerFunc controllers.Func[*capi.GetAllMyRoutineTasksRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetAllMyRoutineTasksRequestDto{}
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

func (b *RoutineTaskBinder) BindCreateRoutineTaskByRoutineId(controllerFunc controllers.Func[*capi.CreateRoutineTaskByRoutineIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateRoutineTaskByRoutineIdRequestDto{}
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

func (b *RoutineTaskBinder) BindUpdateMyRoutineTaskById(controllerFunc controllers.Func[*capi.UpdateMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRoutineTaskByIdRequestDto{}
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

func (b *RoutineTaskBinder) BindHardDeleteMyRoutineTaskById(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTaskByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutineTaskByIdRequestDto{}
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

func (b *RoutineTaskBinder) BindHardDeleteMyRoutineTasksByIds(controllerFunc controllers.Func[*capi.HardDeleteMyRoutineTasksByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.HardDeleteMyRoutineTasksByIdsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		bindRoutineTaskJSON(ctx, requestDto, &requestDto.Body, controllerFunc)
		return
	}
}

func (b *RoutineTaskBinder) BindVisualizeMyRoutineTaskPurposeCount(controllerFunc controllers.Func[*capi.VisualizeMyRoutineTaskPurposeCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.VisualizeMyRoutineTaskPurposeCountRequestDto{}
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
