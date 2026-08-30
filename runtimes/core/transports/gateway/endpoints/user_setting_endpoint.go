package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	userservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/user"
)

type UserSettingEndpointInterface interface {
	GetMySetting(ctx *gin.Context)
	UpdateMySetting(ctx *gin.Context)
}

type UserSettingEndpoint struct {
	userSettingService userservices.UserSettingServiceInterface
}

func NewUserSettingEndpoint(
	userSettingService userservices.UserSettingServiceInterface,
) UserSettingEndpointInterface {
	return &UserSettingEndpoint{
		userSettingService: userSettingService,
	}
}

func (t *UserSettingEndpoint) GetMySetting(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMySettingRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.userSettingService.GetMySetting(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMySettingResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *UserSettingEndpoint) UpdateMySetting(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMySettingRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.userSettingService.UpdateMySetting(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMySettingResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}
