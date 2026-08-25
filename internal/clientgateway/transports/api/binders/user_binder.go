package binders

import (
	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type UserBinderInterface interface {
	BindGetUserData(controllers.Func[*capi.GetUserDataRequestDto]) gin.HandlerFunc
	BindGetMe(controllers.Func[*capi.GetMeRequestDto]) gin.HandlerFunc
	BindUpdateMe(controllers.Func[*capi.UpdateMeRequestDto]) gin.HandlerFunc
}
type UserBinder struct{}

func NewUserBinder() UserBinderInterface { return &UserBinder{} }
func (b *UserBinder) BindGetUserData(controllerFunc controllers.Func[*capi.GetUserDataRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetUserDataRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		controllerFunc(ctx, requestDto)
	}
}
func (b *UserBinder) BindGetMe(controllerFunc controllers.Func[*capi.GetMeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		controllerFunc(ctx, requestDto)
	}
}
func (b *UserBinder) BindUpdateMe(controllerFunc controllers.Func[*capi.UpdateMeRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMeRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("User").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, requestDto)
	}
}
