package binders

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/stations"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/controllers"
)

type StationBinderInterface interface {
	BindGetMyStationById(controllerFunc controllers.Func[*capi.GetMyStationByIdRequestDto]) gin.HandlerFunc
	BindGetAllMyStations(controllerFunc controllers.Func[*capi.GetAllMyStationsRequestDto]) gin.HandlerFunc
	BindCreateStation(controllerFunc controllers.Func[*capi.CreateStationRequestDto]) gin.HandlerFunc
	BindCreateStations(controllerFunc controllers.Func[*capi.CreateStationsRequestDto]) gin.HandlerFunc
	BindUpdateMyStationById(controllerFunc controllers.Func[*capi.UpdateMyStationByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyStationsByIds(controllerFunc controllers.Func[*capi.UpdateMyStationsByIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyStationById(controllerFunc controllers.Func[*capi.RestoreMyStationByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyStationsByIds(controllerFunc controllers.Func[*capi.RestoreMyStationsByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyStationById(controllerFunc controllers.Func[*capi.DeleteMyStationByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyStationsByIds(controllerFunc controllers.Func[*capi.DeleteMyStationsByIdsRequestDto]) gin.HandlerFunc
	BindHardDeleteMyStationById(controllerFunc controllers.Func[*capi.HardDeleteMyStationByIdRequestDto]) gin.HandlerFunc
	BindHardDeleteMyStationsByIds(controllerFunc controllers.Func[*capi.HardDeleteMyStationsByIdsRequestDto]) gin.HandlerFunc

	/* ============================== Visualization Methods ============================== */
	BindVisualizeMyTotalCount(controllerFunc controllers.Func[*capi.VisualizeMyTotalCountRequestDto]) gin.HandlerFunc

	/* ============================== Station Permission Methods ============================== */
	BindGetMyStationPermission(controllerFunc controllers.Func[*capi.GetMyStationPermissionRequestDto]) gin.HandlerFunc
	BindCreateMyStationPermission(controllerFunc controllers.Func[*capi.CreateMyStationPermissionRequestDto]) gin.HandlerFunc
	BindUpsertMyStationPermission(controllerFunc controllers.Func[*capi.UpsertMyStationPermissionRequestDto]) gin.HandlerFunc
	BindUpsertMyStationPermissions(controllerFunc controllers.Func[*capi.UpsertMyStationPermissionsRequestDto]) gin.HandlerFunc
	BindUpdateMyStationPermission(controllerFunc controllers.Func[*capi.UpdateMyStationPermissionRequestDto]) gin.HandlerFunc
	BindTransferMyStationOwnership(controllerFunc controllers.Func[*capi.TransferMyStationOwnershipRequestDto]) gin.HandlerFunc
	BindDeleteMyStationPermission(controllerFunc controllers.Func[*capi.DeleteMyStationPermissionRequestDto]) gin.HandlerFunc
	BindDeleteMyStationPermissions(controllerFunc controllers.Func[*capi.DeleteMyStationPermissionsRequestDto]) gin.HandlerFunc
	BindLeaveMyStation(controllerFunc controllers.Func[*capi.LeaveMyStationRequestDto]) gin.HandlerFunc
	BindLeaveMyStations(controllerFunc controllers.Func[*capi.LeaveMyStationsRequestDto]) gin.HandlerFunc
}

type StationBinder struct{}

func NewStationBinder() StationBinderInterface {
	return &StationBinder{}
}

func (b *StationBinder) BindGetMyStationById(controllerFunc controllers.Func[*capi.GetMyStationByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.GetMyStationByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			isDeleted, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
				return
			}
			request.Param.IsDeleted = &isDeleted
		}

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindGetAllMyStations(controllerFunc controllers.Func[*capi.GetAllMyStationsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.GetAllMyStationsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		areDeletedString := ctx.Query("areDeleted")
		if areDeletedString != "" {
			areDeleted, err := strconv.ParseBool(areDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
				return
			}
			request.Query.AreDeleted = &areDeleted
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindCreateStation(controllerFunc controllers.Func[*capi.CreateStationRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateStationRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindCreateStations(controllerFunc controllers.Func[*capi.CreateStationsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateStationsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindUpdateMyStationById(controllerFunc controllers.Func[*capi.UpdateMyStationByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpdateMyStationByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindUpdateMyStationsByIds(controllerFunc controllers.Func[*capi.UpdateMyStationsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpdateMyStationsByIdsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindRestoreMyStationById(controllerFunc controllers.Func[*capi.RestoreMyStationByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.RestoreMyStationByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Body.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindRestoreMyStationsByIds(controllerFunc controllers.Func[*capi.RestoreMyStationsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.RestoreMyStationsByIdsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindDeleteMyStationById(controllerFunc controllers.Func[*capi.DeleteMyStationByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.DeleteMyStationByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Body.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindDeleteMyStationsByIds(controllerFunc controllers.Func[*capi.DeleteMyStationsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.DeleteMyStationsByIdsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindHardDeleteMyStationById(controllerFunc controllers.Func[*capi.HardDeleteMyStationByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.HardDeleteMyStationByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Body.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindHardDeleteMyStationsByIds(controllerFunc controllers.Func[*capi.HardDeleteMyStationsByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.HardDeleteMyStationsByIdsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Station").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

/* ============================== Visualization Methods ============================== */

func (b *StationBinder) BindVisualizeMyTotalCount(controllerFunc controllers.Func[*capi.VisualizeMyTotalCountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.VisualizeMyTotalCountRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		permissionString := ctx.Query("permission")
		if permissionString == "" {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(fmt.Errorf("permission is required")), ctx)
			return
		}
		request.Query.Permission = permissionString

		controllerFunc(ctx, request)
	}
}

/* ============================== Station Permission Methods ============================== */

func (b *StationBinder) BindGetMyStationPermission(controllerFunc controllers.Func[*capi.GetMyStationPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.GetMyStationPermissionRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.UserPublicId = userPublicId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindCreateMyStationPermission(controllerFunc controllers.Func[*capi.CreateMyStationPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateMyStationPermissionRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindUpsertMyStationPermission(controllerFunc controllers.Func[*capi.UpsertMyStationPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpsertMyStationPermissionRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindUpsertMyStationPermissions(controllerFunc controllers.Func[*capi.UpsertMyStationPermissionsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpsertMyStationPermissionsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindUpdateMyStationPermission(controllerFunc controllers.Func[*capi.UpdateMyStationPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpdateMyStationPermissionRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindTransferMyStationOwnership(controllerFunc controllers.Func[*capi.TransferMyStationOwnershipRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.TransferMyStationOwnershipRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindDeleteMyStationPermission(controllerFunc controllers.Func[*capi.DeleteMyStationPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.DeleteMyStationPermissionRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.UserPublicId = userPublicId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindDeleteMyStationPermissions(controllerFunc controllers.Func[*capi.DeleteMyStationPermissionsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.DeleteMyStationPermissionsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindLeaveMyStation(controllerFunc controllers.Func[*capi.LeaveMyStationRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.LeaveMyStationRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		stationId, err := uuid.Parse(ctx.Param("station-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Station").WithOrigin(err), ctx)
			return
		}
		request.Param.StationId = stationId

		controllerFunc(ctx, request)
	}
}

func (b *StationBinder) BindLeaveMyStations(controllerFunc controllers.Func[*capi.LeaveMyStationsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.LeaveMyStationsRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Station").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}
