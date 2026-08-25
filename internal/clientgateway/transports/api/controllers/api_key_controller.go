package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type APIKeyControllerInterface interface {
	CreateMyAPIKey(*gin.Context, *capi.CreateMyAPIKeyRequestDto)
	ListMyAPIKeys(*gin.Context, *capi.ListMyAPIKeysRequestDto)
	RevokeMyAPIKey(*gin.Context, *capi.RevokeMyAPIKeyRequestDto)
}

type APIKeyController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewAPIKeyController(coreAdapter *coreadapters.CoreAdapter) APIKeyControllerInterface {
	return &APIKeyController{coreAdapter: coreAdapter}
}

func (c *APIKeyController) CreateMyAPIKey(ctx *gin.Context, request *capi.CreateMyAPIKeyRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.CreateMyAPIKeyRequestDto, capi.CreateMyAPIKeyResponseDto](ctx, c.coreAdapter, request, capi.CreateMyAPIKeyOperation, "/core/v1/api-keys/create")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeCreatedClientResponse(ctx, response.Data)
}

func (c *APIKeyController) ListMyAPIKeys(ctx *gin.Context, request *capi.ListMyAPIKeysRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.ListMyAPIKeysRequestDto, capi.ListMyAPIKeysResponseDto](ctx, c.coreAdapter, request, capi.ListMyAPIKeysOperation, "/core/v1/api-keys/list")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *APIKeyController) RevokeMyAPIKey(ctx *gin.Context, request *capi.RevokeMyAPIKeyRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.RevokeMyAPIKeyRequestDto, capi.RevokeMyAPIKeyResponseDto](ctx, c.coreAdapter, request, capi.RevokeMyAPIKeyOperation, "/core/v1/api-keys/revoke")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}
