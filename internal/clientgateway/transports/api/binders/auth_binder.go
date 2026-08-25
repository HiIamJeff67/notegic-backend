package binders

import (
	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type AuthBinderInterface interface {
	BindRegister(controllerFunc controllers.Func[*capi.RegisterRequestDto]) gin.HandlerFunc
	BindRegisterViaGoogle(controllerFunc controllers.Func[*capi.RegisterViaGoogleRequestDto]) gin.HandlerFunc
	BindLogin(controllerFunc controllers.Func[*capi.LoginRequestDto]) gin.HandlerFunc
	BindLoginViaGoogle(controllerFunc controllers.Func[*capi.LoginViaGoogleRequestDto]) gin.HandlerFunc
	BindLogout(controllerFunc controllers.Func[*capi.LogoutRequestDto]) gin.HandlerFunc
	BindSendAuthCode(controllerFunc controllers.Func[*capi.SendAuthCodeRequestDto]) gin.HandlerFunc
	BindValidateEmail(controllerFunc controllers.Func[*capi.ValidateEmailRequestDto]) gin.HandlerFunc
	BindResetEmail(controllerFunc controllers.Func[*capi.ResetEmailRequestDto]) gin.HandlerFunc
	BindForgetPassword(controllerFunc controllers.Func[*capi.ForgetPasswordRequestDto]) gin.HandlerFunc
	BindResetMe(controllerFunc controllers.Func[*capi.ResetMeRequestDto]) gin.HandlerFunc
	BindDeleteMe(controllerFunc controllers.Func[*capi.DeleteMeRequestDto]) gin.HandlerFunc
}

type AuthBinder struct{}

func NewAuthBinder() AuthBinderInterface {
	return &AuthBinder{}
}

func (b *AuthBinder) BindRegister(controllerFunc controllers.Func[*capi.RegisterRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.RegisterRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindRegisterViaGoogle(controllerFunc controllers.Func[*capi.RegisterViaGoogleRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.RegisterViaGoogleRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindLogin(controllerFunc controllers.Func[*capi.LoginRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LoginRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindLoginViaGoogle(controllerFunc controllers.Func[*capi.LoginViaGoogleRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LoginViaGoogleRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindLogout(controllerFunc controllers.Func[*capi.LogoutRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.LogoutRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindSendAuthCode(controllerFunc controllers.Func[*capi.SendAuthCodeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.SendAuthCodeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindValidateEmail(controllerFunc controllers.Func[*capi.ValidateEmailRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.ValidateEmailRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindResetEmail(controllerFunc controllers.Func[*capi.ResetEmailRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.ResetEmailRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindForgetPassword(controllerFunc controllers.Func[*capi.ForgetPasswordRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.ForgetPasswordRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindResetMe(controllerFunc controllers.Func[*capi.ResetMeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.ResetMeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}

func (b *AuthBinder) BindDeleteMe(controllerFunc controllers.Func[*capi.DeleteMeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.DeleteMeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Auth").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}
