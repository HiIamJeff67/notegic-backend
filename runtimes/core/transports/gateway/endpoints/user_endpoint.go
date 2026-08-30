package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	userservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/user"
)

type UserEndpointInterface interface {
	GetUserData(*gin.Context)
	GetMe(*gin.Context)
	UpdateMe(*gin.Context)

	/* ============================== GraphQL Methods ============================== */
	SearchUsers(*gin.Context)
	LoadThemeAuthors(*gin.Context)
}

type UserEndpoint struct {
	userService userservices.UserServiceInterface
}

func NewUserEndpoint(userService userservices.UserServiceInterface) UserEndpointInterface {
	return &UserEndpoint{userService: userService}
}

func (t *UserEndpoint) GetUserData(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetUserDataRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	response, exception := t.userService.GetUserData(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetUserDataResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *response,
	})
}

func (t *UserEndpoint) GetMe(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMeRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	response, exception := t.userService.GetMe(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMeResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *response,
	})
}

func (t *UserEndpoint) UpdateMe(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMeRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	response, exception := t.userService.UpdateMe(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMeResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *response,
	})
}
