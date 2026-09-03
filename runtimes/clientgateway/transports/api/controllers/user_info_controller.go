package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-infos"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type UserInfoControllerInterface interface {
	GetMyInfo(ctx *gin.Context, request *capi.GetMyInfoRequestDto)
	UpdateMyInfo(ctx *gin.Context, request *capi.UpdateMyInfoRequestDto)
}

type UserInfoController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewUserInfoController(
	coreAdapter *coreadapters.CoreAdapter,
) UserInfoControllerInterface {
	return &UserInfoController{
		coreAdapter: coreAdapter,
	}
}

func (c *UserInfoController) GetMyInfo(ctx *gin.Context, request *capi.GetMyInfoRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyInfoRequestDto,
		capi.GetMyInfoResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.GetMyInfoOperation,
		"/core/v1/user-infos/get",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *UserInfoController) UpdateMyInfo(ctx *gin.Context, request *capi.UpdateMyInfoRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyInfoRequestDto,
		capi.UpdateMyInfoResponseDto,
	](
		ctx,
		c.coreAdapter,
		request,
		capi.UpdateMyInfoOperation,
		"/core/v1/user-infos/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}
