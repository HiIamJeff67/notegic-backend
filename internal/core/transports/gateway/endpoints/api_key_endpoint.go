package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	apikeyservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/apikey"
)

type APIKeyEndpointInterface interface {
	CreateMyAPIKey(*gin.Context)
	ListMyAPIKeys(*gin.Context)
	RevokeMyAPIKey(*gin.Context)
}

type APIKeyEndpoint struct {
	service apikeyservices.APIKeyServiceInterface
}

func NewAPIKeyEndpoint(service apikeyservices.APIKeyServiceInterface) APIKeyEndpointInterface {
	return &APIKeyEndpoint{service: service}
}

func (e *APIKeyEndpoint) CreateMyAPIKey(ctx *gin.Context) {
	request := &cgateway.Request[capi.CreateMyAPIKeyRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response, exception := e.service.CreateMyAPIKey(ctx.Request.Context(), &request.Dto)
	writeAPIKeyResponse(ctx, request.Metadata.RequestId, response, exception, http.StatusCreated)
}

func (e *APIKeyEndpoint) ListMyAPIKeys(ctx *gin.Context) {
	request := &cgateway.Request[capi.ListMyAPIKeysRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response, exception := e.service.ListMyAPIKeys(ctx.Request.Context(), &request.Dto)
	writeAPIKeyResponse(ctx, request.Metadata.RequestId, response, exception, http.StatusOK)
}

func (e *APIKeyEndpoint) RevokeMyAPIKey(ctx *gin.Context) {
	request := &cgateway.Request[capi.RevokeMyAPIKeyRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response, exception := e.service.RevokeMyAPIKey(ctx.Request.Context(), &request.Dto)
	writeAPIKeyResponse(ctx, request.Metadata.RequestId, response, exception, http.StatusOK)
}

func writeAPIKeyResponse[T any](
	ctx *gin.Context,
	requestID string,
	data *T,
	exception *cexceptions.Exception,
	status int,
) {
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version:  cgateway.Version,
			Metadata: cgateway.ResponseMetadata{RequestId: requestID, RespondedAt: time.Now()},
			Data:     struct{}{}, Exception: publicException,
		})
		return
	}
	ctx.JSON(status, cgateway.Response[T]{
		Version:  cgateway.Version,
		Metadata: cgateway.ResponseMetadata{RequestId: requestID, RespondedAt: time.Now()},
		Data:     *data,
	})
}
