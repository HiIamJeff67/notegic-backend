package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type RootShelfControllerInterface interface {
	GetMyRootShelfById(ctx *gin.Context, request *capi.GetMyRootShelfByIdRequestDto)
	CreateRootShelf(ctx *gin.Context, request *capi.CreateRootShelfRequestDto)
	CreateRootShelves(ctx *gin.Context, request *capi.CreateRootShelvesRequestDto)
	UpdateMyRootShelfById(ctx *gin.Context, request *capi.UpdateMyRootShelfByIdRequestDto)
	UpdateMyRootShelvesByIds(ctx *gin.Context, reqDto *capi.UpdateMyRootShelvesByIdsRequestDto)
	RestoreMyRootShelfById(ctx *gin.Context, reqDto *capi.RestoreMyRootShelfByIdRequestDto)
	RestoreMyRootShelvesByIds(ctx *gin.Context, reqDto *capi.RestoreMyRootShelvesByIdsRequestDto)
	DeleteMyRootShelfById(ctx *gin.Context, reqDto *capi.DeleteMyRootShelfByIdRequestDto)
	DeleteMyRootShelvesByIds(ctx *gin.Context, reqDto *capi.DeleteMyRootShelvesByIdsRequestDto)

	GetMyRootShelfPermission(ctx *gin.Context, requestDto *capi.GetMyRootShelfPermissionRequestDto)
	CreateMyRootShelfPermission(ctx *gin.Context, requestDto *capi.CreateMyRootShelfPermissionRequestDto)
	UpsertMyRootShelfPermission(ctx *gin.Context, requestDto *capi.UpsertMyRootShelfPermissionRequestDto)
	UpsertMyRootShelfPermissions(ctx *gin.Context, requestDto *capi.UpsertMyRootShelfPermissionsRequestDto)
	UpdateMyRootShelfPermission(ctx *gin.Context, requestDto *capi.UpdateMyRootShelfPermissionRequestDto)
	TransferMyRootShelfOwnership(ctx *gin.Context, requestDto *capi.TransferMyRootShelfOwnershipRequestDto)
	DeleteMyRootShelfPermission(ctx *gin.Context, requestDto *capi.DeleteMyRootShelfPermissionRequestDto)
	DeleteMyRootShelfPermissions(ctx *gin.Context, requestDto *capi.DeleteMyRootShelfPermissionsRequestDto)
	LeaveMyRootShelf(ctx *gin.Context, requestDto *capi.LeaveMyRootShelfRequestDto)
	LeaveMyRootShelves(ctx *gin.Context, requestDto *capi.LeaveMyRootShelvesRequestDto)
}

type RootShelfController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRootShelfController(coreAdapter *coreadapters.CoreAdapter) RootShelfControllerInterface {
	return &RootShelfController{
		coreAdapter: coreAdapter,
	}
}

/* ============================== RootShelf Controller Methods ============================== */

func (c *RootShelfController) GetMyRootShelfById(ctx *gin.Context, request *capi.GetMyRootShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyRootShelfByIdRequestDto,
		capi.GetMyRootShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.GetMyRootShelfByIdOperation,
		"/core/v1/root-shelves/get-by-id",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) CreateRootShelf(ctx *gin.Context, request *capi.CreateRootShelfRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRootShelfRequestDto,
		capi.CreateRootShelfResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.CreateRootShelfOperation,
		"/core/v1/root-shelves/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RootShelfController) CreateRootShelves(ctx *gin.Context, request *capi.CreateRootShelvesRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateRootShelvesRequestDto,
		capi.CreateRootShelvesResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.CreateRootShelvesOperation,
		"/core/v1/root-shelves/create-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RootShelfController) UpdateMyRootShelfById(ctx *gin.Context, request *capi.UpdateMyRootShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyRootShelfByIdRequestDto,
		capi.UpdateMyRootShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpdateMyRootShelfByIdOperation,
		"/core/v1/root-shelves/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) UpdateMyRootShelvesByIds(ctx *gin.Context, requestDto *capi.UpdateMyRootShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyRootShelvesByIdsRequestDto,
		capi.UpdateMyRootShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyRootShelvesByIdsOperation,
		"/core/v1/root-shelves/update-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) RestoreMyRootShelfById(ctx *gin.Context, requestDto *capi.RestoreMyRootShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyRootShelfByIdRequestDto,
		capi.RestoreMyRootShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyRootShelfByIdOperation,
		"/core/v1/root-shelves/restore",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) RestoreMyRootShelvesByIds(ctx *gin.Context, requestDto *capi.RestoreMyRootShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.RestoreMyRootShelvesByIdsRequestDto,
		capi.RestoreMyRootShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RestoreMyRootShelvesByIdsOperation,
		"/core/v1/root-shelves/restore-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) DeleteMyRootShelfById(ctx *gin.Context, requestDto *capi.DeleteMyRootShelfByIdRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyRootShelfByIdRequestDto,
		capi.DeleteMyRootShelfByIdResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRootShelfByIdOperation,
		"/core/v1/root-shelves/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) DeleteMyRootShelvesByIds(ctx *gin.Context, requestDto *capi.DeleteMyRootShelvesByIdsRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.DeleteMyRootShelvesByIdsRequestDto,
		capi.DeleteMyRootShelvesByIdsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRootShelvesByIdsOperation,
		"/core/v1/root-shelves/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) GetMyRootShelfPermission(ctx *gin.Context, requestDto *capi.GetMyRootShelfPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetMyRootShelfPermissionRequestDto, capi.GetMyRootShelfPermissionResponseDto](ctx, c.coreAdapter, requestDto, capi.GetMyRootShelfPermissionOperation, "/core/v1/root-shelves/permissions/get")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) CreateMyRootShelfPermission(ctx *gin.Context, requestDto *capi.CreateMyRootShelfPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.CreateMyRootShelfPermissionRequestDto, capi.CreateMyRootShelfPermissionResponseDto](ctx, c.coreAdapter, requestDto, capi.CreateMyRootShelfPermissionOperation, "/core/v1/root-shelves/permissions/create")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RootShelfController) UpsertMyRootShelfPermission(
	ctx *gin.Context, requestDto *capi.UpsertMyRootShelfPermissionRequestDto,
) {
	response, exception := coreadapters.CallSecurly[capi.UpsertMyRootShelfPermissionRequestDto, capi.UpsertMyRootShelfPermissionResponseDto](ctx, c.coreAdapter, requestDto, capi.UpsertMyRootShelfPermissionOperation, "/core/v1/root-shelves/permissions/upsert")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) UpsertMyRootShelfPermissions(
	ctx *gin.Context, requestDto *capi.UpsertMyRootShelfPermissionsRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpsertMyRootShelfPermissionsRequestDto,
		capi.UpsertMyRootShelfPermissionsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpsertMyRootShelfPermissionsOperation,
		"/core/v1/root-shelves/permissions/upsert-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) UpdateMyRootShelfPermission(ctx *gin.Context, requestDto *capi.UpdateMyRootShelfPermissionRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.UpdateMyRootShelfPermissionRequestDto, capi.UpdateMyRootShelfPermissionResponseDto](ctx, c.coreAdapter, requestDto, capi.UpdateMyRootShelfPermissionOperation, "/core/v1/root-shelves/permissions/update")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) TransferMyRootShelfOwnership(
	ctx *gin.Context, requestDto *capi.TransferMyRootShelfOwnershipRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.TransferMyRootShelfOwnershipRequestDto,
		capi.TransferMyRootShelfOwnershipResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.TransferMyRootShelfOwnershipOperation,
		"/core/v1/root-shelves/ownership/transfer",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *RootShelfController) DeleteMyRootShelfPermission(
	ctx *gin.Context, requestDto *capi.DeleteMyRootShelfPermissionRequestDto,
) {
	_, exception := coreadapters.CallSecurly[
		capi.DeleteMyRootShelfPermissionRequestDto,
		capi.DeleteMyRootShelfPermissionResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRootShelfPermissionOperation,
		"/core/v1/root-shelves/permissions/delete",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *RootShelfController) DeleteMyRootShelfPermissions(
	ctx *gin.Context, requestDto *capi.DeleteMyRootShelfPermissionsRequestDto,
) {
	_, exception := coreadapters.CallSecurly[
		capi.DeleteMyRootShelfPermissionsRequestDto,
		capi.DeleteMyRootShelfPermissionsResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.DeleteMyRootShelfPermissionsOperation,
		"/core/v1/root-shelves/permissions/delete-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *RootShelfController) LeaveMyRootShelf(ctx *gin.Context, requestDto *capi.LeaveMyRootShelfRequestDto) {
	_, exception := coreadapters.CallSecurly[
		capi.LeaveMyRootShelfRequestDto,
		capi.LeaveMyRootShelfResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LeaveMyRootShelfOperation,
		"/core/v1/root-shelves/memberships/leave",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *RootShelfController) LeaveMyRootShelves(ctx *gin.Context, requestDto *capi.LeaveMyRootShelvesRequestDto) {
	_, exception := coreadapters.CallSecurly[
		capi.LeaveMyRootShelvesRequestDto,
		capi.LeaveMyRootShelvesResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LeaveMyRootShelvesOperation,
		"/core/v1/root-shelves/memberships/leave-many",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.Status(http.StatusNoContent)
}
