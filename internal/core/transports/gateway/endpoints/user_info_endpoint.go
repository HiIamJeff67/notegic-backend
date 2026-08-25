package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-infos"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	userservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/user"
)

type UserInfoEndpointInterface interface {
	GetMyInfo(ctx *gin.Context)
	UpdateMyInfo(ctx *gin.Context)

	/* ============================== GraphQL Methods ============================== */
	LoadUserInfos(ctx *gin.Context)
}

type UserInfoEndpoint struct {
	userInfoService userservices.UserInfoServiceInterface
}

func NewUserInfoEndpoint(userInfoService userservices.UserInfoServiceInterface) UserInfoEndpointInterface {
	return &UserInfoEndpoint{userInfoService: userInfoService}
}

func (t *UserInfoEndpoint) GetMyInfo(ctx *gin.Context) {
	request := &cgateway.Request[capi.GetMyInfoRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.userInfoService.GetMyInfo(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.GetMyInfoResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}

func (t *UserInfoEndpoint) UpdateMyInfo(ctx *gin.Context) {
	request := &cgateway.Request[capi.UpdateMyInfoRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.userInfoService.UpdateMyInfo(ctx.Request.Context(), &request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.UpdateMyInfoResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}
