package realtime

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/realtime"
	realtimegatewaycontract "github.com/HiIamJeff67/notegic-backend/contracts/realtime-gateway/v1"
	enumscontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
	yjsworkercontract "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type RealtimeServiceInterface interface {
	CreateMyRealtimeConnectionTicket(ctx context.Context, requestDto *apicontract.CreateMyRealtimeConnectionTicketRequestDto) (*apicontract.CreateMyRealtimeConnectionTicketResponseDto, *exceptions.Exception)
	CreateMyBlockPackChannelTicket(ctx context.Context, requestDto *apicontract.CreateMyBlockPackChannelTicketRequestDto) (*apicontract.CreateMyBlockPackChannelTicketResponseDto, *exceptions.Exception)
}

type RealtimeService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	blockPackRepository repositories.BlockPackRepositoryInterface
}

func NewRealtimeService(
	validator *validator.Validate,
	db *gorm.DB,
	blockPackRepository repositories.BlockPackRepositoryInterface,
) RealtimeServiceInterface {
	return &RealtimeService{
		validator:           validator,
		db:                  db,
		blockPackRepository: blockPackRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (s *RealtimeService) getActorUserPublicId(ctx context.Context) (uuid.UUID, *exceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return uuid.Nil, exception
	}
	var user schemas.User
	result := s.db.WithContext(ctx).
		Model(&schemas.User{}).
		Select("public_id").
		Where("id = ?", actorUserId).
		First(&user)
	if result.Error != nil {
		return uuid.Nil, exceptions.New(
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
	requestDto *apicontract.CreateMyRealtimeConnectionTicketRequestDto,
) (*apicontract.CreateMyRealtimeConnectionTicketResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
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
		RealtimeProtocolVersion: constants.RealtimeProtocolVersion,
	}
	connectionClaims.Subject = userPublicId.String()
	connectionTicket, expiresAt, err := sharedtokens.GenerateRealtimeConnectionTicket(connectionClaims)
	if err != nil {
		return nil, exceptions.New(
			"GenerationFailed",
			"Realtime",
			"CreateMyRealtimeConnectionTicket",
			"Failed to generate the realtime connection ticket",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.CreateMyRealtimeConnectionTicketResponseDto{
		RealtimeEndpoint:        "/" + realtimegatewaycontract.RealtimeDevelopmentBaseURL,
		RealtimeProtocolVersion: constants.RealtimeProtocolVersion,
		ConnectionTicket:        *connectionTicket,
		ExpiresAt:               expiresAt,
	}, nil
}

func (s *RealtimeService) CreateMyBlockPackChannelTicket(
	ctx context.Context,
	requestDto *apicontract.CreateMyBlockPackChannelTicketRequestDto,
) (*apicontract.CreateMyBlockPackChannelTicketResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Block pack channel ticket request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	permission := enumscontract.ChannelPermission(requestDto.Body.Permission)
	sharedAllowedPermissions := permission.AllowedAccessControlPermissions()
	if len(sharedAllowedPermissions) == 0 {
		return nil, exceptions.New(
			"InvalidChannelPermission",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Realtime channel permission is invalid",
			http.StatusBadRequest,
		)
	}
	allowedPermissions := make([]enums.AccessControlPermission, len(sharedAllowedPermissions))
	for index, sharedAllowedPermission := range sharedAllowedPermissions {
		allowedPermissions[index] = enums.AccessControlPermission(sharedAllowedPermission)
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
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return nil, exception
	}

	var yjsDocument schemas.BlockPackYjsDocument
	result := db.
		Where("block_pack_id = ?", blockPack.Id).
		Where("deleted_at IS NULL").
		First(&yjsDocument)
	if result.Error != nil {
		return nil, exceptions.New(
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
		Model(&schemas.BlockPack{}).
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
		return nil, exceptions.New(
			"Unavailable",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Block pack realtime room admission is unavailable",
			http.StatusServiceUnavailable,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 || roomPolicy.MaximumSubscribers <= 0 || roomPolicy.MaximumBlockCount <= 0 {
		return nil, exceptions.New(
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
		RealtimeProtocolVersion:          constants.RealtimeProtocolVersion,
		SchemaVersion:                    yjsworkercontract.YjsBlockPackSchemaVersion,
		RoomAdmissionPolicyVersion:       realtimegatewaycontract.BlockPackRoomAdmissionPolicyVersion,
		RoomAdmissionEnforcementStrategy: string(realtimegatewaycontract.RoomAdmissionEnforcementStrategy_RejectNewSubscriber),
		MaximumSubscribers:               roomPolicy.MaximumSubscribers,
		DocumentQuotaPolicyVersion:       yjsworkercontract.BlockPackDocumentQuotaPolicyVersion,
		MaximumBlockCount:                roomPolicy.MaximumBlockCount,
	}
	channelClaims.Subject = userPublicId.String()
	channelTicket, expiresAt, err := sharedtokens.GenerateRealtimeBlockPackTicket(channelClaims)
	if err != nil {
		return nil, exceptions.New(
			"GenerationFailed",
			"BlockPack",
			"CreateMyBlockPackChannelTicket",
			"Failed to generate the block pack channel ticket",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.CreateMyBlockPackChannelTicketResponseDto{
		ChannelTicket:              *channelTicket,
		ExpiresAt:                  expiresAt,
		ChannelType:                "BlockPack",
		ChannelId:                  blockPack.Id,
		Permission:                 string(permission),
		RoomName:                   fmt.Sprintf("%s:%s", yjsworkercontract.YjsBlockPackRoomPrefix, blockPack.Id),
		FragmentName:               yjsworkercontract.YjsBlockPackFragmentName,
		SchemaId:                   yjsworkercontract.YjsBlockPackSchemaId,
		SchemaVersion:              yjsworkercontract.YjsBlockPackSchemaVersion,
		RealtimeProtocolVersion:    constants.RealtimeProtocolVersion,
		DocumentQuotaPolicyVersion: yjsworkercontract.BlockPackDocumentQuotaPolicyVersion,
		MaximumBlockCount:          roomPolicy.MaximumBlockCount,
		LastUpdateSequence:         yjsDocument.LastUpdateSequence,
		CompactedUntilSequence:     yjsDocument.CompactedUntilSequence,
	}, nil
}
