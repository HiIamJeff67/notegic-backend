package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/stations"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type StationControllerInterface interface {
	GetMyStationById(ctx *gin.Context, request *capi.GetMyStationByIdRequestDto)
	GetAllMyStations(ctx *gin.Context, request *capi.GetAllMyStationsRequestDto)
	CreateStation(ctx *gin.Context, request *capi.CreateStationRequestDto)
	CreateStations(ctx *gin.Context, request *capi.CreateStationsRequestDto)
	UpdateMyStationById(ctx *gin.Context, request *capi.UpdateMyStationByIdRequestDto)
	UpdateMyStationsByIds(ctx *gin.Context, request *capi.UpdateMyStationsByIdsRequestDto)
	RestoreMyStationById(ctx *gin.Context, request *capi.RestoreMyStationByIdRequestDto)
	RestoreMyStationsByIds(ctx *gin.Context, request *capi.RestoreMyStationsByIdsRequestDto)
	DeleteMyStationById(ctx *gin.Context, request *capi.DeleteMyStationByIdRequestDto)
	DeleteMyStationsByIds(ctx *gin.Context, request *capi.DeleteMyStationsByIdsRequestDto)
	HardDeleteMyStationById(ctx *gin.Context, request *capi.HardDeleteMyStationByIdRequestDto)
	HardDeleteMyStationsByIds(ctx *gin.Context, request *capi.HardDeleteMyStationsByIdsRequestDto)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyTotalCount(ctx *gin.Context, request *capi.VisualizeMyTotalCountRequestDto)

	/* ============================== Station Permission Methods ============================== */
	GetMyStationPermission(ctx *gin.Context, request *capi.GetMyStationPermissionRequestDto)
	CreateMyStationPermission(ctx *gin.Context, request *capi.CreateMyStationPermissionRequestDto)
	UpsertMyStationPermission(ctx *gin.Context, request *capi.UpsertMyStationPermissionRequestDto)
	UpsertMyStationPermissions(ctx *gin.Context, request *capi.UpsertMyStationPermissionsRequestDto)
	UpdateMyStationPermission(ctx *gin.Context, request *capi.UpdateMyStationPermissionRequestDto)
	TransferMyStationOwnership(ctx *gin.Context, request *capi.TransferMyStationOwnershipRequestDto)
	DeleteMyStationPermission(ctx *gin.Context, request *capi.DeleteMyStationPermissionRequestDto)
	DeleteMyStationPermissions(ctx *gin.Context, request *capi.DeleteMyStationPermissionsRequestDto)
	LeaveMyStation(ctx *gin.Context, request *capi.LeaveMyStationRequestDto)
	LeaveMyStations(ctx *gin.Context, request *capi.LeaveMyStationsRequestDto)
}

type StationController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewStationController(coreAdapter *coreadapters.CoreAdapter) StationControllerInterface {
	return &StationController{
		coreAdapter: coreAdapter,
	}
}

func (c *StationController) GetMyStationById(ctx *gin.Context, request *capi.GetMyStationByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyStationByIdRequestDto,
		capi.GetMyStationByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.GetMyStationByIdOperation,
		"/core/v1/stations/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) GetAllMyStations(ctx *gin.Context, request *capi.GetAllMyStationsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetAllMyStationsRequestDto,
		capi.GetAllMyStationsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.GetAllMyStationsOperation,
		"/core/v1/stations/get-all",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) CreateStation(ctx *gin.Context, request *capi.CreateStationRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateStationRequestDto,
		capi.CreateStationResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.CreateStationOperation,
		"/core/v1/stations/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *StationController) CreateStations(ctx *gin.Context, request *capi.CreateStationsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateStationsRequestDto,
		capi.CreateStationsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.CreateStationsOperation,
		"/core/v1/stations/create-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *StationController) UpdateMyStationById(ctx *gin.Context, request *capi.UpdateMyStationByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyStationByIdRequestDto,
		capi.UpdateMyStationByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpdateMyStationByIdOperation,
		"/core/v1/stations/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) UpdateMyStationsByIds(ctx *gin.Context, request *capi.UpdateMyStationsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyStationsByIdsRequestDto,
		capi.UpdateMyStationsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpdateMyStationsByIdsOperation,
		"/core/v1/stations/update-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) RestoreMyStationById(ctx *gin.Context, request *capi.RestoreMyStationByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyStationByIdRequestDto,
		capi.RestoreMyStationByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.RestoreMyStationByIdOperation,
		"/core/v1/stations/restore",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) RestoreMyStationsByIds(ctx *gin.Context, request *capi.RestoreMyStationsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyStationsByIdsRequestDto,
		capi.RestoreMyStationsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.RestoreMyStationsByIdsOperation,
		"/core/v1/stations/restore-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) DeleteMyStationById(ctx *gin.Context, request *capi.DeleteMyStationByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyStationByIdRequestDto,
		capi.DeleteMyStationByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.DeleteMyStationByIdOperation,
		"/core/v1/stations/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) DeleteMyStationsByIds(ctx *gin.Context, request *capi.DeleteMyStationsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyStationsByIdsRequestDto,
		capi.DeleteMyStationsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.DeleteMyStationsByIdsOperation,
		"/core/v1/stations/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) HardDeleteMyStationById(ctx *gin.Context, request *capi.HardDeleteMyStationByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.HardDeleteMyStationByIdRequestDto,
		capi.HardDeleteMyStationByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.HardDeleteMyStationByIdOperation,
		"/core/v1/stations/hard-delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) HardDeleteMyStationsByIds(ctx *gin.Context, request *capi.HardDeleteMyStationsByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.HardDeleteMyStationsByIdsRequestDto,
		capi.HardDeleteMyStationsByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.HardDeleteMyStationsByIdsOperation,
		"/core/v1/stations/hard-delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

/* ============================== Controller Methods for Visualization ============================== */

func (c *StationController) VisualizeMyTotalCount(ctx *gin.Context, request *capi.VisualizeMyTotalCountRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.VisualizeMyTotalCountRequestDto,
		capi.VisualizeMyTotalCountResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.VisualizeMyTotalCountOperation,
		"/core/v1/stations/visualizations/total-count",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

/* ============================== Controller Methods for Station Permissions ============================== */

func (c *StationController) GetMyStationPermission(ctx *gin.Context, request *capi.GetMyStationPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyStationPermissionRequestDto,
		capi.GetMyStationPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.GetMyStationPermissionOperation,
		"/core/v1/stations/permissions/get",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) CreateMyStationPermission(ctx *gin.Context, request *capi.CreateMyStationPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateMyStationPermissionRequestDto,
		capi.CreateMyStationPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.CreateMyStationPermissionOperation,
		"/core/v1/stations/permissions/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *StationController) UpsertMyStationPermission(
	ctx *gin.Context, request *capi.UpsertMyStationPermissionRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpsertMyStationPermissionRequestDto,
		capi.UpsertMyStationPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpsertMyStationPermissionOperation,
		"/core/v1/stations/permissions/upsert",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) UpsertMyStationPermissions(
	ctx *gin.Context, request *capi.UpsertMyStationPermissionsRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpsertMyStationPermissionsRequestDto,
		capi.UpsertMyStationPermissionsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpsertMyStationPermissionsOperation,
		"/core/v1/stations/permissions/upsert-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) UpdateMyStationPermission(ctx *gin.Context, request *capi.UpdateMyStationPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyStationPermissionRequestDto,
		capi.UpdateMyStationPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpdateMyStationPermissionOperation,
		"/core/v1/stations/permissions/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) TransferMyStationOwnership(
	ctx *gin.Context, request *capi.TransferMyStationOwnershipRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.TransferMyStationOwnershipRequestDto,
		capi.TransferMyStationOwnershipResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.TransferMyStationOwnershipOperation,
		"/core/v1/stations/ownership/transfer",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *StationController) DeleteMyStationPermission(
	ctx *gin.Context, request *capi.DeleteMyStationPermissionRequestDto,
) {
	_, exception := coreadapters.CallSecurly[
		capi.DeleteMyStationPermissionRequestDto,
		capi.DeleteMyStationPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.DeleteMyStationPermissionOperation,
		"/core/v1/stations/permissions/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, capi.DeleteMyStationPermissionResponseDto{})
}

func (c *StationController) DeleteMyStationPermissions(
	ctx *gin.Context, request *capi.DeleteMyStationPermissionsRequestDto,
) {
	_, exception := coreadapters.CallSecurly[
		capi.DeleteMyStationPermissionsRequestDto,
		capi.DeleteMyStationPermissionsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.DeleteMyStationPermissionsOperation,
		"/core/v1/stations/permissions/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *StationController) LeaveMyStation(ctx *gin.Context, request *capi.LeaveMyStationRequestDto) {
	_, exception := coreadapters.CallSecurly[
		capi.LeaveMyStationRequestDto,
		capi.LeaveMyStationResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.LeaveMyStationOperation,
		"/core/v1/stations/memberships/leave",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *StationController) LeaveMyStations(ctx *gin.Context, request *capi.LeaveMyStationsRequestDto) {
	_, exception := coreadapters.CallSecurly[
		capi.LeaveMyStationsRequestDto,
		capi.LeaveMyStationsResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.LeaveMyStationsOperation,
		"/core/v1/stations/memberships/leave-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}
