package binders

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/realtime"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type RealtimeBinderInterface interface {
	BindCreateMyRealtimeConnectionTicket(controllerFunc controllers.Func[*capi.CreateMyRealtimeConnectionTicketRequestDto]) gin.HandlerFunc
	BindCreateMyBlockPackChannelTicket(controllerFunc controllers.Func[*capi.CreateMyBlockPackChannelTicketRequestDto]) gin.HandlerFunc
}

type RealtimeBinder struct{}

func NewRealtimeBinder() RealtimeBinderInterface {
	return &RealtimeBinder{}
}

func (b *RealtimeBinder) BindCreateMyRealtimeConnectionTicket(controllerFunc controllers.Func[*capi.CreateMyRealtimeConnectionTicketRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateMyRealtimeConnectionTicketRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		controllerFunc(ctx, requestDto)
	}
}

func (b *RealtimeBinder) BindCreateMyBlockPackChannelTicket(controllerFunc controllers.Func[*capi.CreateMyBlockPackChannelTicketRequestDto]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &capi.CreateMyBlockPackChannelTicketRequestDto{}
		requestDto.Header.UserAgent = ctx.GetHeader("User-Agent")

		if err := ctx.ShouldBindJSON(&requestDto.Body); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("BlockPack").WithOrigin(err), ctx)
			return
		}

		controllerFunc(ctx, requestDto)
	}
}
