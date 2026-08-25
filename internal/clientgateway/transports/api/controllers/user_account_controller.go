package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-accounts"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type UserAccountControllerInterface interface {
	GetMyAccount(ctx *gin.Context, requestDto *capi.GetMyAccountRequestDto)
	UpdateMyAccount(ctx *gin.Context, requestDto *capi.UpdateMyAccountRequestDto)
	BindGoogleAccount(ctx *gin.Context, requestDto *capi.BindGoogleAccountRequestDto)
	UnbindGoogleAccount(ctx *gin.Context, requestDto *capi.UnbindGoogleAccountRequestDto)
}

type UserAccountController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewUserAccountController(coreAdapter *coreadapters.CoreAdapter) UserAccountControllerInterface {
	return &UserAccountController{
		coreAdapter: coreAdapter,
	}
}

func (c *UserAccountController) GetMyAccount(
	ctx *gin.Context,
	requestDto *capi.GetMyAccountRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.GetMyAccountRequestDto,
		capi.GetMyAccountResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.GetMyAccountOperation,
		"/core/v1/user-accounts/get",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *UserAccountController) UpdateMyAccount(
	ctx *gin.Context,
	requestDto *capi.UpdateMyAccountRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UpdateMyAccountRequestDto,
		capi.UpdateMyAccountResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UpdateMyAccountOperation,
		"/core/v1/user-accounts/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *UserAccountController) BindGoogleAccount(
	ctx *gin.Context,
	requestDto *capi.BindGoogleAccountRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.BindGoogleAccountRequestDto,
		capi.BindGoogleAccountResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.BindGoogleAccountOperation,
		"/core/v1/user-accounts/google/bind",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *UserAccountController) UnbindGoogleAccount(
	ctx *gin.Context,
	requestDto *capi.UnbindGoogleAccountRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.UnbindGoogleAccountRequestDto,
		capi.UnbindGoogleAccountResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.UnbindGoogleAccountOperation,
		"/core/v1/user-accounts/google/unbind",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
