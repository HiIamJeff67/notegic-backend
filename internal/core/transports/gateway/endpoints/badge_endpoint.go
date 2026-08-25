package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/badges"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	otherservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/other"
)

type BadgeEndpointInterface interface {
	LoadUserBadges(ctx *gin.Context)
}

type BadgeEndpoint struct {
	badgeService otherservices.BadgeServiceInterface
}

func NewBadgeEndpoint(
	badgeService otherservices.BadgeServiceInterface,
) BadgeEndpointInterface {
	return &BadgeEndpoint{
		badgeService: badgeService,
	}
}

func (t *BadgeEndpoint) LoadUserBadges(ctx *gin.Context) {
	request := &cgateway.Request[capi.LoadUserBadgesRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDtos, exception := t.badgeService.GetPublicBadgesByUserPublicIds(ctx.Request.Context(), request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.LoadUserBadgesResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: responseDtos,
	})
}
