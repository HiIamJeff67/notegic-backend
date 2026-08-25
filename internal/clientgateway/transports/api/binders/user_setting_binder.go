package binders

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type UserSettingBinderInterface interface {
	BindGetMySetting(controllerFunc controllers.Func[*capi.GetMySettingRequestDto]) gin.HandlerFunc
	BindUpdateMySetting(controllerFunc controllers.Func[*capi.UpdateMySettingRequestDto]) gin.HandlerFunc
}

type UserSettingBinder struct{}

func NewUserSettingBinder() UserSettingBinderInterface {
	return &UserSettingBinder{}
}

func (b *UserSettingBinder) BindGetMySetting(controllerFunc controllers.Func[*capi.GetMySettingRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.GetMySettingRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		controllerFunc(ctx, requestDto)
	}
}

func (b *UserSettingBinder) BindUpdateMySetting(controllerFunc controllers.Func[*capi.UpdateMySettingRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.UpdateMySettingRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("UserSetting").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}
