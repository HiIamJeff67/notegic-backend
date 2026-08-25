package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	shelfservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/shelves"
)

type RootShelfEndpointInterface interface {
	GetMyRootShelfById(ctx *gin.Context)
	CreateRootShelf(ctx *gin.Context)
	CreateRootShelves(ctx *gin.Context)
	UpdateMyRootShelfById(ctx *gin.Context)
	UpdateMyRootShelvesByIds(ctx *gin.Context)
	RestoreMyRootShelfById(ctx *gin.Context)
	RestoreMyRootShelvesByIds(ctx *gin.Context)
	DeleteMyRootShelfById(ctx *gin.Context)
	DeleteMyRootShelvesByIds(ctx *gin.Context)
	GetMyRootShelfPermission(ctx *gin.Context)
	CreateMyRootShelfPermission(ctx *gin.Context)
	UpsertMyRootShelfPermission(ctx *gin.Context)
	UpsertMyRootShelfPermissions(ctx *gin.Context)
	UpdateMyRootShelfPermission(ctx *gin.Context)
	TransferMyRootShelfOwnership(ctx *gin.Context)
	DeleteMyRootShelfPermission(ctx *gin.Context)
	DeleteMyRootShelfPermissions(ctx *gin.Context)
	LeaveMyRootShelf(ctx *gin.Context)
	LeaveMyRootShelves(ctx *gin.Context)

	/* ============================== GraphQL Methods ============================== */
	SearchRootShelves(ctx *gin.Context)
}

type RootShelfEndpoint struct {
	rootShelfService shelfservices.RootShelfServiceInterface
}

func NewRootShelfEndpoint(
	service shelfservices.RootShelfServiceInterface,
) RootShelfEndpointInterface {
	return &RootShelfEndpoint{
		rootShelfService: service,
	}
}

/* ============================== RootShelf Endpoint Methods ============================== */

func (t *RootShelfEndpoint) GetMyRootShelfById(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRootShelfByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.GetMyRootShelfById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRootShelfByIdResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) CreateRootShelf(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRootShelfRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.CreateRootShelf(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateRootShelfResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) CreateRootShelves(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateRootShelvesRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.CreateRootShelves(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateRootShelvesResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) UpdateMyRootShelfById(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRootShelfByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.UpdateMyRootShelfById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRootShelfByIdResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) UpdateMyRootShelvesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRootShelvesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.UpdateMyRootShelvesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRootShelvesByIdsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) RestoreMyRootShelfById(ctx *gin.Context) {
	request := &cgateway.Request[capi.RestoreMyRootShelfByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.RestoreMyRootShelfById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.RestoreMyRootShelfByIdResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) RestoreMyRootShelvesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.RestoreMyRootShelvesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.RestoreMyRootShelvesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.RestoreMyRootShelvesByIdsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) DeleteMyRootShelfById(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRootShelfByIdRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.DeleteMyRootShelfById(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRootShelfByIdResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) DeleteMyRootShelvesByIds(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRootShelvesByIdsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.DeleteMyRootShelvesByIds(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRootShelvesByIdsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) GetMyRootShelfPermission(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyRootShelfPermissionRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.GetMyRootShelfPermission(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyRootShelfPermissionResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RootShelfEndpoint) CreateMyRootShelfPermission(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateMyRootShelfPermissionRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.CreateMyRootShelfPermission(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.CreateMyRootShelfPermissionResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RootShelfEndpoint) UpsertMyRootShelfPermission(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpsertMyRootShelfPermissionRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.UpsertMyRootShelfPermission(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpsertMyRootShelfPermissionResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RootShelfEndpoint) UpsertMyRootShelfPermissions(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpsertMyRootShelfPermissionsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.UpsertMyRootShelfPermissions(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpsertMyRootShelfPermissionsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) UpdateMyRootShelfPermission(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyRootShelfPermissionRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.UpdateMyRootShelfPermission(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyRootShelfPermissionResponseDto]{Version: cgateway.Version, Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()}, Data: *responseDto})
}

func (t *RootShelfEndpoint) TransferMyRootShelfOwnership(ctx *gin.Context) {
	request := &cgateway.Request[capi.TransferMyRootShelfOwnershipRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.TransferMyRootShelfOwnership(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.TransferMyRootShelfOwnershipResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) DeleteMyRootShelfPermission(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRootShelfPermissionRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.DeleteMyRootShelfPermission(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRootShelfPermissionResponseDto]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: request.Metadata.RequestId, RespondedAt: time.Now()},
		Data:     *responseDto,
	})
}

func (t *RootShelfEndpoint) DeleteMyRootShelfPermissions(ctx *gin.Context) {
	request := &cgateway.Request[capi.DeleteMyRootShelfPermissionsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.rootShelfService.DeleteMyRootShelfPermissions(ctx.Request.Context(), &request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.DeleteMyRootShelfPermissionsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *RootShelfEndpoint) LeaveMyRootShelf(ctx *gin.Context) {
	request := &cgateway.Request[capi.LeaveMyRootShelfRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if exception := t.rootShelfService.LeaveMyRootShelf(ctx.Request.Context(), &request.Dto); exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.LeaveMyRootShelfResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: capi.LeaveMyRootShelfResponseDto{},
	})
}

func (t *RootShelfEndpoint) LeaveMyRootShelves(ctx *gin.Context) {
	request := &cgateway.Request[capi.LeaveMyRootShelvesRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if exception := t.rootShelfService.LeaveMyRootShelves(ctx.Request.Context(), &request.Dto); exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.LeaveMyRootShelvesResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: capi.LeaveMyRootShelvesResponseDto{},
	})
}
