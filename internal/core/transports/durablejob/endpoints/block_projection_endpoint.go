package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"
	cdurablejobdto "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	blockservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/blocks"
)

type BlockProjectionEndpoint struct {
	blockService blockservices.BlockServiceInterface
}

func NewBlockProjectionEndpoint(blockService blockservices.BlockServiceInterface) BlockProjectionEndpoint {
	return BlockProjectionEndpoint{blockService: blockService}
}

func (e BlockProjectionEndpoint) Apply(ctx *gin.Context) {
	request := &cgateway.Request[cdurablejobdto.ApplyBlockProjectionRequestDto]{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		exception := cexceptions.New(
			"InvalidRequest",
			"DurableJob",
			cdurablejobdto.ApplyBlockProjectionOperation,
			"The DurableJob projection request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
		ctx.JSON(exception.HTTPStatusCode(), cgateway.Response[cdurablejobdto.ApplyBlockProjectionResponseDto]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Exception: exception,
		})
		return
	}

	documents := make([]capi.ApplyBlockProjectionDocumentRequestDto, len(request.Dto.Documents))
	for index, document := range request.Dto.Documents {
		documents[index] = capi.ApplyBlockProjectionDocumentRequestDto{
			BlockPackId: document.BlockPackId,
			Projection: capi.ApplyBlockProjectionRequestDto{
				SchemaId:          document.Projection.SchemaId,
				SchemaVersion:     document.Projection.SchemaVersion,
				ProjectedSequence: document.Projection.ProjectedSequence,
				Blocks:            document.Projection.Blocks,
			},
		}
	}

	responseDto, err := e.blockService.ApplyMany(
		ctx.Request.Context(),
		documents,
	)
	if err != nil {
		exception := cexceptions.New(
			"FailedToApplyProjection",
			"DurableJob",
			cdurablejobdto.ApplyBlockProjectionOperation,
			"Failed to apply projected blocks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
		ctx.JSON(exception.HTTPStatusCode(), cgateway.Response[cdurablejobdto.ApplyBlockProjectionResponseDto]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   request.Metadata.RequestId,
				RespondedAt: time.Now(),
			},
			Exception: exception,
		})
		return
	}

	ctx.JSON(http.StatusOK, cgateway.Response[cdurablejobdto.ApplyBlockProjectionResponseDto]{
		Version: cgateway.Version,
		Metadata: cgateway.ResponseMetadata{
			RequestId:   request.Metadata.RequestId,
			RespondedAt: time.Now(),
		},
		Data: cdurablejobdto.ApplyBlockProjectionResponseDto{
			AppliedBlockPackIds: []uuid.UUID(responseDto),
		},
	})
}
