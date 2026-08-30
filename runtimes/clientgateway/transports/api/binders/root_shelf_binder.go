package binders

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type RootShelfBinderInterface interface {
	BindGetMyRootShelfById(controllerFunc controllers.Func[*capi.GetMyRootShelfByIdRequestDto]) gin.HandlerFunc
	BindCreateRootShelf(controllerFunc controllers.Func[*capi.CreateRootShelfRequestDto]) gin.HandlerFunc
	BindCreateRootShelves(controllerFunc controllers.Func[*capi.CreateRootShelvesRequestDto]) gin.HandlerFunc
	BindUpdateMyRootShelfById(controllerFunc controllers.Func[*capi.UpdateMyRootShelfByIdRequestDto]) gin.HandlerFunc
	BindUpdateMyRootShelvesByIds(controllerFunc controllers.Func[*capi.UpdateMyRootShelvesByIdsRequestDto]) gin.HandlerFunc
	BindRestoreMyRootShelfById(controllerFunc controllers.Func[*capi.RestoreMyRootShelfByIdRequestDto]) gin.HandlerFunc
	BindRestoreMyRootShelvesByIds(controllerFunc controllers.Func[*capi.RestoreMyRootShelvesByIdsRequestDto]) gin.HandlerFunc
	BindDeleteMyRootShelfById(controllerFunc controllers.Func[*capi.DeleteMyRootShelfByIdRequestDto]) gin.HandlerFunc
	BindDeleteMyRootShelvesByIds(controllerFunc controllers.Func[*capi.DeleteMyRootShelvesByIdsRequestDto]) gin.HandlerFunc
	BindGetMyRootShelfPermission(controllerFunc controllers.Func[*capi.GetMyRootShelfPermissionRequestDto]) gin.HandlerFunc
	BindCreateMyRootShelfPermission(controllerFunc controllers.Func[*capi.CreateMyRootShelfPermissionRequestDto]) gin.HandlerFunc
	BindUpsertMyRootShelfPermission(controllerFunc controllers.Func[*capi.UpsertMyRootShelfPermissionRequestDto]) gin.HandlerFunc
	BindUpsertMyRootShelfPermissions(controllerFunc controllers.Func[*capi.UpsertMyRootShelfPermissionsRequestDto]) gin.HandlerFunc
	BindUpdateMyRootShelfPermission(controllerFunc controllers.Func[*capi.UpdateMyRootShelfPermissionRequestDto]) gin.HandlerFunc
	BindTransferMyRootShelfOwnership(controllerFunc controllers.Func[*capi.TransferMyRootShelfOwnershipRequestDto]) gin.HandlerFunc
	BindDeleteMyRootShelfPermission(controllerFunc controllers.Func[*capi.DeleteMyRootShelfPermissionRequestDto]) gin.HandlerFunc
	BindDeleteMyRootShelfPermissions(controllerFunc controllers.Func[*capi.DeleteMyRootShelfPermissionsRequestDto]) gin.HandlerFunc
	BindLeaveMyRootShelf(controllerFunc controllers.Func[*capi.LeaveMyRootShelfRequestDto]) gin.HandlerFunc
	BindLeaveMyRootShelves(controllerFunc controllers.Func[*capi.LeaveMyRootShelvesRequestDto]) gin.HandlerFunc
}

type RootShelfBinder struct{}

func NewRootShelfBinder() RootShelfBinderInterface {
	return &RootShelfBinder{}
}

func (b *RootShelfBinder) BindGetMyRootShelfById(controllerFunc controllers.Func[*capi.GetMyRootShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.GetMyRootShelfByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")

		isDeletedString := ctx.Query("isDeleted")
		if isDeletedString != "" {
			isDeleted, err := strconv.ParseBool(isDeletedString)
			if err != nil {
				sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
				return
			}
			request.Param.IsDeleted = &isDeleted
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		request.Param.RootShelfId = rootShelfId

		controllerFunc(ctx, request)
	}
}

func (b *RootShelfBinder) BindCreateRootShelf(controllerFunc controllers.Func[*capi.CreateRootShelfRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateRootShelfRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *RootShelfBinder) BindCreateRootShelves(controllerFunc controllers.Func[*capi.CreateRootShelvesRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateRootShelvesRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}

func (b *RootShelfBinder) BindUpdateMyRootShelfById(controllerFunc controllers.Func[*capi.UpdateMyRootShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpdateMyRootShelfByIdRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		request.Param.RootShelfId = rootShelfId

		controllerFunc(ctx, request)
	}
}

func (b *RootShelfBinder) BindUpdateMyRootShelvesByIds(controllerFunc controllers.Func[*capi.UpdateMyRootShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqDto capi.UpdateMyRootShelvesByIdsRequestDto

		reqDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&reqDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &reqDto)
	}
}

func (b *RootShelfBinder) BindRestoreMyRootShelfById(controllerFunc controllers.Func[*capi.RestoreMyRootShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqDto capi.RestoreMyRootShelfByIdRequestDto

		reqDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&reqDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		reqDto.Body.RootShelfId = rootShelfId

		controllerFunc(ctx, &reqDto)
	}
}

func (b *RootShelfBinder) BindRestoreMyRootShelvesByIds(controllerFunc controllers.Func[*capi.RestoreMyRootShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqDto capi.RestoreMyRootShelvesByIdsRequestDto

		reqDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&reqDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &reqDto)
	}
}

func (b *RootShelfBinder) BindDeleteMyRootShelfById(controllerFunc controllers.Func[*capi.DeleteMyRootShelfByIdRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqDto capi.DeleteMyRootShelfByIdRequestDto

		reqDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&reqDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		reqDto.Body.RootShelfId = rootShelfId

		controllerFunc(ctx, &reqDto)
	}
}

func (b *RootShelfBinder) BindDeleteMyRootShelvesByIds(controllerFunc controllers.Func[*capi.DeleteMyRootShelvesByIdsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqDto capi.DeleteMyRootShelvesByIdsRequestDto

		reqDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&reqDto.Body); err != nil {
			exception := cexceptions.InvalidDto("Shelf").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, &reqDto)
	}
}

func (b *RootShelfBinder) BindGetMyRootShelfPermission(controllerFunc controllers.Func[*capi.GetMyRootShelfPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyRootShelfPermissionRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.UserPublicId = userPublicId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindCreateMyRootShelfPermission(controllerFunc controllers.Func[*capi.CreateMyRootShelfPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateMyRootShelfPermissionRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindUpsertMyRootShelfPermission(controllerFunc controllers.Func[*capi.UpsertMyRootShelfPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpsertMyRootShelfPermissionRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindUpsertMyRootShelfPermissions(controllerFunc controllers.Func[*capi.UpsertMyRootShelfPermissionsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpsertMyRootShelfPermissionsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindUpdateMyRootShelfPermission(controllerFunc controllers.Func[*capi.UpdateMyRootShelfPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyRootShelfPermissionRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.UserPublicId = userPublicId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindTransferMyRootShelfOwnership(controllerFunc controllers.Func[*capi.TransferMyRootShelfOwnershipRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.TransferMyRootShelfOwnershipRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindDeleteMyRootShelfPermission(controllerFunc controllers.Func[*capi.DeleteMyRootShelfPermissionRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteMyRootShelfPermissionRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		userPublicId, err := uuid.Parse(ctx.Param("user-public-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.UserPublicId = userPublicId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindDeleteMyRootShelfPermissions(controllerFunc controllers.Func[*capi.DeleteMyRootShelfPermissionsRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteMyRootShelfPermissionsRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindLeaveMyRootShelf(controllerFunc controllers.Func[*capi.LeaveMyRootShelfRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LeaveMyRootShelfRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		rootShelfId, err := uuid.Parse(ctx.Param("root-shelf-id"))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("Shelf").WithOrigin(err), ctx)
			return
		}
		requestDto.Param.RootShelfId = rootShelfId

		controllerFunc(ctx, requestDto)
	}
}

func (b *RootShelfBinder) BindLeaveMyRootShelves(controllerFunc controllers.Func[*capi.LeaveMyRootShelvesRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LeaveMyRootShelvesRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Shelf").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, requestDto)
	}
}
