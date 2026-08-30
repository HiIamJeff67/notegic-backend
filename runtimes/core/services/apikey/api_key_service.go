package apikey

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apikeycache "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/redis/apikey"
)

type APIKeyServiceInterface interface {
	CreateMyAPIKey(context.Context, *capi.CreateMyAPIKeyRequestDto) (*capi.CreateMyAPIKeyResponseDto, *cexceptions.Exception)
	ListMyAPIKeys(context.Context, *capi.ListMyAPIKeysRequestDto) (*capi.ListMyAPIKeysResponseDto, *cexceptions.Exception)
	RevokeMyAPIKey(context.Context, *capi.RevokeMyAPIKeyRequestDto) (*capi.RevokeMyAPIKeyResponseDto, *cexceptions.Exception)
}

type APIKeyService struct {
	validator  *validator.Validate
	db         *gorm.DB
	repository srepositories.APIKeyRepositoryInterface
	cache      *apikeycache.APIKeyCacheClient
}

func NewAPIKeyService(
	validator *validator.Validate,
	db *gorm.DB,
	repository srepositories.APIKeyRepositoryInterface,
	cache ...*apikeycache.APIKeyCacheClient,
) APIKeyServiceInterface {
	var cacheClient *apikeycache.APIKeyCacheClient
	if len(cache) > 0 {
		cacheClient = cache[0]
	}
	return &APIKeyService{validator: validator, db: db, repository: repository, cache: cacheClient}
}

func (s *APIKeyService) CreateMyAPIKey(
	ctx context.Context,
	request *capi.CreateMyAPIKeyRequestDto,
) (*capi.CreateMyAPIKeyResponseDto, *cexceptions.Exception) {
	if exception := s.validate(request, "CreateMyAPIKey"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	secret, keyPrefix, keyHash, err := sharedtokens.GenerateAPIKey()
	if err != nil {
		return nil, cexceptions.New("APIKeyCreateFailed", "APIKey", "CreateMyAPIKey", "The API key could not be generated", http.StatusInternalServerError, true).WithOrigin(err)
	}
	now := time.Now()
	created, exception := s.repository.Create(&sschemas.APIKey{
		Id: uuid.New(), PublicId: uuid.New(), UserId: userId,
		Name: request.Body.Name, KeyPrefix: keyPrefix, KeyHash: keyHash,
		ExpiresAt: request.Body.ExpiresAt, CreatedAt: now, UpdatedAt: now,
	}, srepositories.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	return &capi.CreateMyAPIKeyResponseDto{
		PublicId: created.PublicId.String(), Name: created.Name, KeyPrefix: created.KeyPrefix,
		Secret: secret, ExpiresAt: created.ExpiresAt, CreatedAt: created.CreatedAt,
	}, nil
}

func (s *APIKeyService) ListMyAPIKeys(
	ctx context.Context,
	request *capi.ListMyAPIKeysRequestDto,
) (*capi.ListMyAPIKeysResponseDto, *cexceptions.Exception) {
	if exception := s.validate(request, "ListMyAPIKeys"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	keys, exception := s.repository.GetAllByUserId(userId, srepositories.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	items := make([]capi.APIKeySummary, 0, len(keys))
	for _, key := range keys {
		items = append(items, capi.APIKeySummary{
			PublicId: key.PublicId.String(), Name: key.Name, KeyPrefix: key.KeyPrefix,
			LastUsedAt: key.LastUsedAt, ExpiresAt: key.ExpiresAt, RevokedAt: key.RevokedAt, CreatedAt: key.CreatedAt,
		})
	}
	return &capi.ListMyAPIKeysResponseDto{Items: items}, nil
}

func (s *APIKeyService) RevokeMyAPIKey(
	ctx context.Context,
	request *capi.RevokeMyAPIKeyRequestDto,
) (*capi.RevokeMyAPIKeyResponseDto, *cexceptions.Exception) {
	if exception := s.validate(request, "RevokeMyAPIKey"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	publicId, err := uuid.Parse(request.Param.PublicId)
	if err != nil {
		return nil, cexceptions.InvalidInput("APIKey").WithOrigin(err)
	}
	keys, exception := s.repository.GetAllByUserId(userId, srepositories.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	var key *sschemas.APIKey
	for index := range keys {
		if keys[index].PublicId == publicId {
			key = &keys[index]
			break
		}
	}
	if key == nil {
		return nil, cexceptions.New("APIKeyNotFound", "APIKey", "RevokeMyAPIKey", "The API key was not found", http.StatusNotFound)
	}
	now := time.Now()
	if exception := s.repository.Revoke(key.Id, now, srepositories.WithDB(s.db.WithContext(ctx))); exception != nil {
		return nil, exception
	}
	if s.cache != nil {
		_ = s.cache.Delete(key.KeyHash)
	}
	return &capi.RevokeMyAPIKeyResponseDto{RevokedAt: now.Format(time.RFC3339Nano)}, nil
}

func (s *APIKeyService) validate(request any, operation string) *cexceptions.Exception {
	if err := s.validator.Struct(request); err != nil {
		return cexceptions.New("InvalidRequest", "APIKey", operation, "API key request is invalid", http.StatusBadRequest).WithOrigin(err)
	}
	return nil
}
