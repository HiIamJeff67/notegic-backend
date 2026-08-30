package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/themes"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	otherservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/other"
)

type ThemeEndpointInterface interface {
	SearchThemes(ctx *gin.Context)
}

type ThemeEndpoint struct {
	themeService otherservices.ThemeServiceInterface
}

func NewThemeEndpoint(
	themeService otherservices.ThemeServiceInterface,
) ThemeEndpointInterface {
	return &ThemeEndpoint{
		themeService: themeService,
	}
}

func (t *ThemeEndpoint) SearchThemes(ctx *gin.Context) {
	request := &cgateway.Request[capi.SearchThemesRequestDto]{}
	if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	responseDto, exception := t.themeService.SearchPublicThemes(ctx.Request.Context(), request.Dto)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.SearchThemesResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}
