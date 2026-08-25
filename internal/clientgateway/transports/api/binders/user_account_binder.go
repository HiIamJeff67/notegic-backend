package binders

import (
	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-accounts"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type UserAccountBinderInterface interface {
	BindGetMyAccount(controllerFunc controllers.Func[*capi.GetMyAccountRequestDto]) gin.HandlerFunc
	BindUpdateMyAccount(controllerFunc controllers.Func[*capi.UpdateMyAccountRequestDto]) gin.HandlerFunc
	BindBindGoogleAccount(controllerFunc controllers.Func[*capi.BindGoogleAccountRequestDto]) gin.HandlerFunc
	BindUnbindGoogleAccount(controllerFunc controllers.Func[*capi.UnbindGoogleAccountRequestDto]) gin.HandlerFunc
}

type UserAccountBinder struct{}

func NewUserAccountBinder() UserAccountBinderInterface {
	return &UserAccountBinder{}
}

func (b *UserAccountBinder) BindGetMyAccount(controllerFunc controllers.Func[*capi.GetMyAccountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMyAccountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		controllerFunc(ctx, requestDto)
	}
}

func (b *UserAccountBinder) BindUpdateMyAccount(controllerFunc controllers.Func[*capi.UpdateMyAccountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMyAccountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("UserAccount").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *UserAccountBinder) BindBindGoogleAccount(controllerFunc controllers.Func[*capi.BindGoogleAccountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.BindGoogleAccountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("UserAccount").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *UserAccountBinder) BindUnbindGoogleAccount(controllerFunc controllers.Func[*capi.UnbindGoogleAccountRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UnbindGoogleAccountRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("UserAccount").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}
