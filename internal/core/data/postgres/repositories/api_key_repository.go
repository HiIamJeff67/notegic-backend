package repositories

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
)

type APIKeyRepositoryInterface interface {
	GetOneByKeyHash(keyHash string, opts ...options.RepositoryOptions) (*schemas.APIKey, *exceptions.Exception)
	GetAllByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.APIKey, *exceptions.Exception)
	Create(apiKey *schemas.APIKey, opts ...options.RepositoryOptions) (*schemas.APIKey, *exceptions.Exception)
	MarkUsed(id uuid.UUID, usedAt time.Time, opts ...options.RepositoryOptions) *exceptions.Exception
	Revoke(id uuid.UUID, revokedAt time.Time, opts ...options.RepositoryOptions) *exceptions.Exception
}

type APIKeyRepository struct{}

func NewAPIKeyRepository() APIKeyRepositoryInterface {
	return &APIKeyRepository{}
}

func (r *APIKeyRepository) GetOneByKeyHash(
	keyHash string,
	opts ...options.RepositoryOptions,
) (*schemas.APIKey, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	apiKey := &schemas.APIKey{}
	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("key_hash = ?", keyHash).
		First(apiKey)
	if result.Error != nil || apiKey.Id == uuid.Nil {
		return nil, exceptions.New(
			"APIKeyNotFound",
			"Repository",
			"GetOneByKeyHash",
			"The API key was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	return apiKey, nil
}

func (r *APIKeyRepository) GetAllByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.APIKey, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	apiKeys := []schemas.APIKey{}
	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(&apiKeys)
	if result.Error != nil {
		return nil, exceptions.New(
			"APIKeyListFailed",
			"Repository",
			"GetAllByUserId",
			"The API keys could not be loaded",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return apiKeys, nil
}

func (r *APIKeyRepository) Create(
	apiKey *schemas.APIKey,
	opts ...options.RepositoryOptions,
) (*schemas.APIKey, *exceptions.Exception) {
	if apiKey == nil {
		return nil, exceptions.New(
			"APIKeyRequired",
			"Repository",
			"Create",
			"The API key is required",
			http.StatusBadRequest,
		)
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Create(apiKey)
	if result.Error != nil {
		return nil, exceptions.New(
			"APIKeyCreateFailed",
			"Repository",
			"Create",
			"The API key could not be created",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return apiKey, nil
}

func (r *APIKeyRepository) MarkUsed(
	id uuid.UUID,
	usedAt time.Time,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("last_used_at", usedAt)
	if result.Error != nil {
		return exceptions.New(
			"APIKeyUsageUpdateFailed",
			"Repository",
			"MarkUsed",
			"The API key usage timestamp could not be updated",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return nil
}

func (r *APIKeyRepository) Revoke(
	id uuid.UUID,
	revokedAt time.Time,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", revokedAt)
	if result.Error != nil {
		return exceptions.New(
			"APIKeyRevokeFailed",
			"Repository",
			"Revoke",
			"The API key could not be revoked",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return nil
}
