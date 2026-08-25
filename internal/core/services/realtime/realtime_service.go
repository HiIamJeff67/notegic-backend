package realtime

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/realtime"
	crealtimegateway "github.com/HiIamJeff67/notegic-backend/contracts/realtime-gateway/v1"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
)

type RealtimeServiceInterface interface {
	CreateMyRealtimeConnectionTicket(ctx context.Context, requestDto *capi.CreateMyRealtimeConnectionTicketRequestDto) (*capi.CreateMyRealtimeConnectionTicketResponseDto, *cexceptions.Exception)
	CreateMyBlockPackChannelTicket(ctx context.Context, requestDto *capi.CreateMyBlockPackChannelTicketRequestDto) (*capi.CreateMyBlockPackChannelTicketResponseDto, *cexceptions.Exception)
}

type RealtimeService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	blockPackRepository srepositories.BlockPackRepositoryInterface
}

func NewRealtimeService(
	validator *validator.Validate,
	db *gorm.DB,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
) RealtimeServiceInterface {
	return &RealtimeService{
		validator:           validator,
		db:                  db,
		blockPackRepository: blockPackRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (s *RealtimeService) getActorUserPublicId(ctx context.Context) (uuid.UUID, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return uuid.Nil, exception
	}
	var user sschemas.User
	result := s.db.WithContext(ctx).
		Model(&sschemas.User{}).
		Select("public_id").
		Where("id = ?", actorUserId).
		First(&user)
	if result.Error != nil {
		return uuid.Nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveActor",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	return user.PublicId, nil
}

/* ============================== Service Methods for Realtime ============================== */

func (s *RealtimeService) CreateMyRealtimeConnectionTicket(
	ctx context.Context,
	requestDto *capi.CreateMyRealtimeConnectionTicketRequestDto,
) (*capi.CreateMyRealtimeConnectionTicketResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"Realtime",
			"CreateMyRealtimeConnectionTicket",
			"Realtime connection ticket request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	userPublicId, exception := s.getActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}
	userAgentHash := sha256.Sum256([]byte(requestDto.Header.UserAgent))
	connectionClaims := sharedtokens.RealtimeConnectionTicketClaims{
		UserAgentHash:           fmt.Sprintf("%x", userAgentHash),
		RealtimeProtocolVersion: sconstants.RealtimeProtocolVersion,
	}
	connectionClaims.Subject = userPublicId.String()
	connectionTicket, expiresAt, err := sharedtokens.GenerateRealtimeConnectionTicket(connectionClaims)
	if err != nil {
		return nil, cexceptions.New(
			"GenerationFailed",
			"Realtime",
			"CreateMyRealtimeConnectionTicket",
			"Failed to generate the realtime connection ticket",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.CreateMyRealtimeConnectionTicketResponseDto{
		RealtimeEndpoint:        "/" + crealtimegateway.RealtimeDevelopmentBaseURL,
		RealtimeProtocolVersion: sconstants.RealtimeProtocolVersion,
		ConnectionTicket:        *connectionTicket,
		ExpiresAt:               expiresAt,
	}, nil
}

func (s *RealtimeService) CreateMyBlockPackChannelTicket(
	ctx context.Context,
	requestDto *capi.CreateMyBlockPackChannelTicketRequestDto,
) (*capi.CreateMyBlockPackChannelTicketResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Block pack channel ticket request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	permission := cenums.ChannelPermission(requestDto.Body.Permission)
	sharedAllowedPermissions := permission.AllowedAccessControlPermissions()
	if len(sharedAllowedPermissions) == 0 {
		return nil, cexceptions.New(
			"InvalidChannelPermission",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Realtime channel permission is invalid",
			http.StatusBadRequest,
		)
	}
	allowedPermissions := make([]cenums.AccessControlPermission, len(sharedAllowedPermissions))
	for index, sharedAllowedPermission := range sharedAllowedPermissions {
		allowedPermissions[index] = cenums.AccessControlPermission(sharedAllowedPermission)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	userPublicId, exception := s.getActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}

	blockPack, exception := s.blockPackRepository.CheckPermissionAndGetOneById(
		requestDto.Body.BlockPackId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return nil, exception
	}

	var yjsDocument sschemas.BlockPackYjsDocument
	result := db.
		Where("block_pack_id = ?", blockPack.Id).
		Where("deleted_at IS NULL").
		First(&yjsDocument)
	if result.Error != nil {
		return nil, cexceptions.New(
			"NotFound",
			"BlockPackYjsDocument",
			"CreateMyBlockPackChannelTicket",
			"Block pack document was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	var roomPolicy struct {
		MaximumSubscribers int32
		MaximumBlockCount  int32
	}
	result = db.
		Model(&sschemas.BlockPack{}).
		Select(`
			"PlanLimitationTable".max_realtime_room_subscriber_count AS maximum_subscribers,
			"PlanLimitationTable".max_block_count_per_block_pack AS maximum_block_count
		`).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Joins(`INNER JOIN "RootShelfTable" ON "RootShelfTable".id = "SubShelfTable".root_shelf_id`).
		Joins(`INNER JOIN "UserTable" ON "UserTable".id = "RootShelfTable".owner_id`).
		Joins(`INNER JOIN "PlanLimitationTable" ON "PlanLimitationTable".key = "UserTable".plan`).
		Where(`"BlockPackTable".id = ?`, blockPack.Id).
		Where(`"BlockPackTable".deleted_at IS NULL`).
		Where(`"SubShelfTable".deleted_at IS NULL`).
		Where(`"RootShelfTable".deleted_at IS NULL`).
		Scan(&roomPolicy)
	if result.Error != nil {
		return nil, cexceptions.New(
			"Unavailable",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Block pack realtime room admission is unavailable",
			http.StatusServiceUnavailable,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 || roomPolicy.MaximumSubscribers <= 0 || roomPolicy.MaximumBlockCount <= 0 {
		return nil, cexceptions.New(
			"Unavailable",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Block pack realtime room admission is unavailable",
			http.StatusServiceUnavailable,
		).WithOrigin(gorm.ErrRecordNotFound)
	}

	userAgentHash := sha256.Sum256([]byte(requestDto.Header.UserAgent))
	channelClaims := sharedtokens.RealtimeBlockPackTicketClaims{
		UserAgentHash:                    fmt.Sprintf("%x", userAgentHash),
		ChannelType:                      "BlockPack",
		ChannelId:                        blockPack.Id.String(),
		Permission:                       string(permission),
		RealtimeProtocolVersion:          sconstants.RealtimeProtocolVersion,
		SchemaVersion:                    cyjsworker.YjsBlockPackSchemaVersion,
		RoomAdmissionPolicyVersion:       crealtimegateway.BlockPackRoomAdmissionPolicyVersion,
		RoomAdmissionEnforcementStrategy: string(crealtimegateway.RoomAdmissionEnforcementStrategy_RejectNewSubscriber),
		MaximumSubscribers:               roomPolicy.MaximumSubscribers,
		DocumentQuotaPolicyVersion:       cyjsworker.BlockPackDocumentQuotaPolicyVersion,
		MaximumBlockCount:                roomPolicy.MaximumBlockCount,
	}
	channelClaims.Subject = userPublicId.String()
	channelTicket, expiresAt, err := sharedtokens.GenerateRealtimeBlockPackTicket(channelClaims)
	if err != nil {
		return nil, cexceptions.New(
			"GenerationFailed",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Failed to generate the block pack channel ticket",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.CreateMyBlockPackChannelTicketResponseDto{
		ChannelTicket:              *channelTicket,
		ExpiresAt:                  expiresAt,
		ChannelType:                "BlockPack",
		ChannelId:                  blockPack.Id,
		Permission:                 string(permission),
		RoomName:                   fmt.Sprintf("%s:%s", cyjsworker.YjsBlockPackRoomPrefix, blockPack.Id),
		FragmentName:               cyjsworker.YjsBlockPackFragmentName,
		SchemaId:                   cyjsworker.YjsBlockPackSchemaId,
		SchemaVersion:              cyjsworker.YjsBlockPackSchemaVersion,
		RealtimeProtocolVersion:    sconstants.RealtimeProtocolVersion,
		DocumentQuotaPolicyVersion: cyjsworker.BlockPackDocumentQuotaPolicyVersion,
		MaximumBlockCount:          roomPolicy.MaximumBlockCount,
		LastUpdateSequence:         yjsDocument.LastUpdateSequence,
		CompactedUntilSequence:     yjsDocument.CompactedUntilSequence,
	}, nil
}
