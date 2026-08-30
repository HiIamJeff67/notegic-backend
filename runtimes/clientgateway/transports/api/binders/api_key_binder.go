package binders

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
)

type APIKeyBinderInterface interface {
	BindCreateMyAPIKey(controllers.Func[*capi.CreateMyAPIKeyRequestDto]) gin.HandlerFunc
	BindListMyAPIKeys(controllers.Func[*capi.ListMyAPIKeysRequestDto]) gin.HandlerFunc
	BindRevokeMyAPIKey(controllers.Func[*capi.RevokeMyAPIKeyRequestDto]) gin.HandlerFunc
}

type APIKeyBinder struct{}

func NewAPIKeyBinder() APIKeyBinderInterface { return &APIKeyBinder{} }

func (b *APIKeyBinder) BindCreateMyAPIKey(controllerFunc controllers.Func[*capi.CreateMyAPIKeyRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.CreateMyAPIKeyRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		if err := ctx.ShouldBindJSON(&request.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("APIKey").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, request)
	}
}

func (b *APIKeyBinder) BindListMyAPIKeys(controllerFunc controllers.Func[*capi.ListMyAPIKeysRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.ListMyAPIKeysRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		controllerFunc(ctx, request)
	}
}

func (b *APIKeyBinder) BindRevokeMyAPIKey(controllerFunc controllers.Func[*capi.RevokeMyAPIKeyRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := &capi.RevokeMyAPIKeyRequestDto{}
		request.Header.UserAgent = ctx.GetHeader("User-Agent")
		request.Param.PublicId = ctx.Param("api-key-id")
		controllerFunc(ctx, request)
	}
}
