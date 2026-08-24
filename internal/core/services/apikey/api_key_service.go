package apikey

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"
	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	apikeycache "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/apikey"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

type APIKeyServiceInterface interface {
	CreateMyAPIKey(context.Context, *apicontract.CreateMyAPIKeyRequestDto) (*apicontract.CreateMyAPIKeyResponseDto, *exceptions.Exception)
	ListMyAPIKeys(context.Context, *apicontract.ListMyAPIKeysRequestDto) (*apicontract.ListMyAPIKeysResponseDto, *exceptions.Exception)
	RevokeMyAPIKey(context.Context, *apicontract.RevokeMyAPIKeyRequestDto) (*apicontract.RevokeMyAPIKeyResponseDto, *exceptions.Exception)
}

type APIKeyService struct {
	validator  *validator.Validate
	db         *gorm.DB
	repository repositories.APIKeyRepositoryInterface
	cache      *apikeycache.APIKeyCacheClient
}

func NewAPIKeyService(
	validator *validator.Validate,
	db *gorm.DB,
	repository repositories.APIKeyRepositoryInterface,
	cache ...*apikeycache.APIKeyCacheClient,
) APIKeyServiceInterface {
	if db == nil {
		db = data.DB
	}
	var cacheClient *apikeycache.APIKeyCacheClient
	if len(cache) > 0 {
		cacheClient = cache[0]
	}
	return &APIKeyService{validator: validator, db: db, repository: repository, cache: cacheClient}
}

func (s *APIKeyService) CreateMyAPIKey(
	ctx context.Context,
	request *apicontract.CreateMyAPIKeyRequestDto,
) (*apicontract.CreateMyAPIKeyResponseDto, *exceptions.Exception) {
	if exception := s.validate(request, "CreateMyAPIKey"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	secret, keyPrefix, keyHash, err := sharedtokens.GenerateAPIKey()
	if err != nil {
		return nil, exceptions.New("APIKeyCreateFailed", "APIKey", "CreateMyAPIKey", "The API key could not be generated", http.StatusInternalServerError, true).WithOrigin(err)
	}
	now := time.Now()
	created, exception := s.repository.Create(&schemas.APIKey{
		Id: uuid.New(), PublicId: uuid.New(), UserId: userId,
		Name: request.Body.Name, KeyPrefix: keyPrefix, KeyHash: keyHash,
		ExpiresAt: request.Body.ExpiresAt, CreatedAt: now, UpdatedAt: now,
	}, options.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	return &apicontract.CreateMyAPIKeyResponseDto{
		PublicId: created.PublicId.String(), Name: created.Name, KeyPrefix: created.KeyPrefix,
		Secret: secret, ExpiresAt: created.ExpiresAt, CreatedAt: created.CreatedAt,
	}, nil
}

func (s *APIKeyService) ListMyAPIKeys(
	ctx context.Context,
	request *apicontract.ListMyAPIKeysRequestDto,
) (*apicontract.ListMyAPIKeysResponseDto, *exceptions.Exception) {
	if exception := s.validate(request, "ListMyAPIKeys"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	keys, exception := s.repository.GetAllByUserId(userId, options.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	items := make([]apicontract.APIKeySummary, 0, len(keys))
	for _, key := range keys {
		items = append(items, apicontract.APIKeySummary{
			PublicId: key.PublicId.String(), Name: key.Name, KeyPrefix: key.KeyPrefix,
			LastUsedAt: key.LastUsedAt, ExpiresAt: key.ExpiresAt, RevokedAt: key.RevokedAt, CreatedAt: key.CreatedAt,
		})
	}
	return &apicontract.ListMyAPIKeysResponseDto{Items: items}, nil
}

func (s *APIKeyService) RevokeMyAPIKey(
	ctx context.Context,
	request *apicontract.RevokeMyAPIKeyRequestDto,
) (*apicontract.RevokeMyAPIKeyResponseDto, *exceptions.Exception) {
	if exception := s.validate(request, "RevokeMyAPIKey"); exception != nil {
		return nil, exception
	}
	userId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	publicId, err := uuid.Parse(request.Param.PublicId)
	if err != nil {
		return nil, exceptions.InvalidInput("APIKey").WithOrigin(err)
	}
	keys, exception := s.repository.GetAllByUserId(userId, options.WithDB(s.db.WithContext(ctx)))
	if exception != nil {
		return nil, exception
	}
	var key *schemas.APIKey
	for index := range keys {
		if keys[index].PublicId == publicId {
			key = &keys[index]
			break
		}
	}
	if key == nil {
		return nil, exceptions.New("APIKeyNotFound", "APIKey", "RevokeMyAPIKey", "The API key was not found", http.StatusNotFound)
	}
	now := time.Now()
	if exception := s.repository.Revoke(key.Id, now, options.WithDB(s.db.WithContext(ctx))); exception != nil {
		return nil, exception
	}
	if s.cache != nil {
		_ = s.cache.Delete(key.KeyHash)
	}
	return &apicontract.RevokeMyAPIKeyResponseDto{RevokedAt: now.Format(time.RFC3339Nano)}, nil
}

func (s *APIKeyService) validate(request any, operation string) *exceptions.Exception {
	if err := s.validator.Struct(request); err != nil {
		return exceptions.New("InvalidRequest", "APIKey", operation, "API key request is invalid", http.StatusBadRequest).WithOrigin(err)
	}
	return nil
}
