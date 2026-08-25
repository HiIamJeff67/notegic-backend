package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/items"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	shelfservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/shelves"
)

type ItemEndpointInterface interface {
	SearchItems(ctx *gin.Context)
}

type ItemEndpoint struct {
	itemService shelfservices.ItemServiceInterface
}

func NewItemEndpoint(
	itemService shelfservices.ItemServiceInterface,
) ItemEndpointInterface {
	return &ItemEndpoint{
		itemService: itemService,
	}
}

func (t *ItemEndpoint) SearchItems(ctx *gin.Context) {
	request := &cgateway.Request[capi.SearchItemsRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userId, exception := contexts.GetActorUserId(ctx.Request.Context())
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	responseDto, exception := t.itemService.SearchPrivateItems(ctx.Request.Context(), userId, request.Dto)
	if exception != nil {
		publicException := exception.ToPublic()
		ctx.JSON(publicException.HTTPStatusCode(), cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Data:      struct{}{},
			Exception: publicException,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[capi.SearchItemsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}
