package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	crealtimegateway "github.com/HiIamJeff67/notegic-backend/contracts/realtime-gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	realtimelease "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
)

type BlockPackEndpoint struct {
	realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient
}

func NewBlockPackEndpoint(realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient) BlockPackEndpoint {
	return BlockPackEndpoint{
		realtimeLeaseCache: realtimeLeaseCache,
	}
}

func (e BlockPackEndpoint) GetParticipants(ctx *gin.Context) {
	requestId := ctx.GetHeader("X-Request-Id")
	if requestId == "" {
		requestId = uuid.NewString()
	}
	blockPackId, err := uuid.Parse(ctx.Param("block-pack-id"))
	if err != nil || blockPackId == uuid.Nil {
		exception := cexceptions.New(
			"InvalidRequest",
			"RealtimeGateway",
			crealtimegateway.GetBlockPackParticipantsOperation,
			"The RealtimeGateway participant request is invalid",
			http.StatusBadRequest,
		)
		if err != nil {
			exception = exception.WithOrigin(err)
		}
		ctx.JSON(exception.HTTPStatusCode(), crealtimegateway.Response[crealtimegateway.GetBlockPackParticipantsResponseDto]{
			Version: crealtimegateway.Version,
			Metadata: crealtimegateway.ResponseMetadata{
				RequestId:   requestId,
				RespondedAt: time.Now(),
			},
			Data: crealtimegateway.GetBlockPackParticipantsResponseDto{
				Participants: []crealtimegateway.BlockPackParticipantResponseDto{},
			},
			Exception: exception,
		})
		return
	}
	if e.realtimeLeaseCache == nil {
		exception := cexceptions.New(
			"RealtimeLeaseCacheRequired",
			"RealtimeGateway",
			crealtimegateway.GetBlockPackParticipantsOperation,
			"The RealtimeGateway lease cache is unavailable",
			http.StatusServiceUnavailable,
			true,
		)
		ctx.JSON(exception.HTTPStatusCode(), crealtimegateway.Response[crealtimegateway.GetBlockPackParticipantsResponseDto]{
			Version: crealtimegateway.Version,
			Metadata: crealtimegateway.ResponseMetadata{
				RequestId:   requestId,
				RespondedAt: time.Now(),
			},
			Data: crealtimegateway.GetBlockPackParticipantsResponseDto{
				Participants: []crealtimegateway.BlockPackParticipantResponseDto{},
			},
			Exception: exception,
		})
		return
	}

	participants, err := e.realtimeLeaseCache.GetBlockPackParticipants(blockPackId)
	if err != nil {
		exception := cexceptions.New(
			"RealtimePresenceUnavailable",
			"RealtimeGateway",
			crealtimegateway.GetBlockPackParticipantsOperation,
			"Realtime participant presence is unavailable",
			http.StatusServiceUnavailable,
			true,
		).WithOrigin(err)
		ctx.JSON(exception.HTTPStatusCode(), crealtimegateway.Response[crealtimegateway.GetBlockPackParticipantsResponseDto]{
			Version: crealtimegateway.Version,
			Metadata: crealtimegateway.ResponseMetadata{
				RequestId:   requestId,
				RespondedAt: time.Now(),
			},
			Data: crealtimegateway.GetBlockPackParticipantsResponseDto{
				Participants: []crealtimegateway.BlockPackParticipantResponseDto{},
			},
			Exception: exception,
		})
		return
	}

	responseDto := crealtimegateway.GetBlockPackParticipantsResponseDto{
		Participants: make([]crealtimegateway.BlockPackParticipantResponseDto, len(participants)),
	}
	for index, participant := range participants {
		responseDto.Participants[index] = crealtimegateway.BlockPackParticipantResponseDto{
			UserPublicId:      participant.UserPublicId,
			ChannelPermission: participant.ChannelPermission,
			ConnectionCount:   participant.ConnectionCount,
		}
	}

	ctx.JSON(http.StatusOK, crealtimegateway.Response[crealtimegateway.GetBlockPackParticipantsResponseDto]{
		Version: crealtimegateway.Version,
		Metadata: crealtimegateway.ResponseMetadata{
			RequestId:   requestId,
			RespondedAt: time.Now(),
		},
		Data: responseDto,
	})
}
