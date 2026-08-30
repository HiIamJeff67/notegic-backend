package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type UserControllerInterface interface {
	GetUserData(ctx *gin.Context, requestDto *capi.GetUserDataRequestDto)
	GetMe(ctx *gin.Context, requestDto *capi.GetMeRequestDto)
	UpdateMe(ctx *gin.Context, requestDto *capi.UpdateMeRequestDto)
}
type UserController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewUserController(coreAdapter *coreadapters.CoreAdapter) UserControllerInterface {
	return &UserController{
		coreAdapter: coreAdapter,
	}
}

func (c *UserController) GetUserData(ctx *gin.Context, requestDto *capi.GetUserDataRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetUserDataRequestDto, capi.GetUserDataResponseDto](
		ctx, c.coreAdapter, requestDto, capi.GetUserDataOperation, "/core/v1/users/data",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *UserController) GetMe(ctx *gin.Context, requestDto *capi.GetMeRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.GetMeRequestDto, capi.GetMeResponseDto](
		ctx, c.coreAdapter, requestDto, capi.GetMeOperation, "/core/v1/users/me",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *UserController) UpdateMe(ctx *gin.Context, requestDto *capi.UpdateMeRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.UpdateMeRequestDto, capi.UpdateMeResponseDto](
		ctx, c.coreAdapter, requestDto, capi.UpdateMeOperation, "/core/v1/users/me/update",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}
