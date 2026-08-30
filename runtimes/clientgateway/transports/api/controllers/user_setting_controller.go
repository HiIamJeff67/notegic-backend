package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type UserSettingControllerInterface interface {
	GetMySetting(ctx *gin.Context, requestDto *capi.GetMySettingRequestDto)
	UpdateMySetting(ctx *gin.Context, requestDto *capi.UpdateMySettingRequestDto)
}

type UserSettingController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewUserSettingController(coreAdapter *coreadapters.CoreAdapter) UserSettingControllerInterface {
	return &UserSettingController{
		coreAdapter: coreAdapter,
	}
}

func (c *UserSettingController) GetMySetting(ctx *gin.Context, requestDto *capi.GetMySettingRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMySettingRequestDto,
		capi.GetMySettingResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMySettingOperation,
		"/core/v1/user-settings/get",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *UserSettingController) UpdateMySetting(ctx *gin.Context, requestDto *capi.UpdateMySettingRequestDto) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMySettingRequestDto,
		capi.UpdateMySettingResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMySettingOperation,
		"/core/v1/user-settings/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
