package binders

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-infos"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type UserInfoBinderInterface interface {
	BindGetMyInfo(controllerFunc controllers.Func[*capi.GetMyInfoRequestDto]) gin.HandlerFunc
	BindUpdateMyInfo(controllerFunc controllers.Func[*capi.UpdateMyInfoRequestDto]) gin.HandlerFunc
}

type UserInfoBinder struct{}

func NewUserInfoBinder() UserInfoBinderInterface {
	return &UserInfoBinder{}
}

func (b *UserInfoBinder) BindGetMyInfo(controllerFunc controllers.Func[*capi.GetMyInfoRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.GetMyInfoRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		controllerFunc(ctx, request)
	}
}

func (b *UserInfoBinder) BindUpdateMyInfo(controllerFunc controllers.Func[*capi.UpdateMyInfoRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.UpdateMyInfoRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			exception := cexceptions.InvalidDto("UserInfo").WithOrigin(err)
			sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
			return
		}

		controllerFunc(ctx, request)
	}
}
