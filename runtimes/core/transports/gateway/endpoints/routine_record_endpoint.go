package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-records"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
)

type RoutineRecordEndpointInterface interface {
	SearchRoutineRecords(ctx *gin.Context)
}

type RoutineRecordEndpoint struct {
	routineRecordService routineservices.RoutineRecordServiceInterface
}

func NewRoutineRecordEndpoint(
	routineRecordService routineservices.RoutineRecordServiceInterface,
) RoutineRecordEndpointInterface {
	return &RoutineRecordEndpoint{
		routineRecordService: routineRecordService,
	}
}

func (t *RoutineRecordEndpoint) SearchRoutineRecords(ctx *gin.Context) {
	request := &cgateway.Request[capi.SearchRoutineRecordsRequestDto]{}
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

	responseDto, exception := t.routineRecordService.SearchPrivateRoutineRecords(
		ctx.Request.Context(),
		userId,
		request.Dto,
	)
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

	ctx.JSON(http.StatusOK, cgateway.Response[capi.SearchRoutineRecordsResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: *responseDto,
	})
}
