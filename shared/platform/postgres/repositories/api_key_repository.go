package repositories

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type APIKeyRepositoryInterface interface {
	GetOneByKeyHash(keyHash string, opts ...RepositoryOptions) (*schemas.APIKey, *cexceptions.Exception)
	GetAllByUserId(userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.APIKey, *cexceptions.Exception)
	Create(apiKey *schemas.APIKey, opts ...RepositoryOptions) (*schemas.APIKey, *cexceptions.Exception)
	MarkUsed(id uuid.UUID, usedAt time.Time, opts ...RepositoryOptions) *cexceptions.Exception
	Revoke(id uuid.UUID, revokedAt time.Time, opts ...RepositoryOptions) *cexceptions.Exception
}

type APIKeyRepository struct {
	db         *gorm.DB
	exceptions exceptions.APIKeyException
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepositoryInterface {
	return &APIKeyRepository{
		db: db, exceptions: exceptions.NewAPIKeyException()}
}

func (r *APIKeyRepository) GetOneByKeyHash(
	keyHash string,
	opts ...RepositoryOptions,
) (*schemas.APIKey, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	apiKey := &schemas.APIKey{}
	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("key_hash = ?", keyHash).
		First(apiKey)
	if result.Error != nil || apiKey.Id == uuid.Nil {
		return nil, r.exceptions.NewForDomain(
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
	opts ...RepositoryOptions,
) ([]schemas.APIKey, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	apiKeys := []schemas.APIKey{}
	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(&apiKeys)
	if result.Error != nil {
		return nil, r.exceptions.NewForDomain(
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
	opts ...RepositoryOptions,
) (*schemas.APIKey, *cexceptions.Exception) {
	if apiKey == nil {
		return nil, r.exceptions.NewForDomain(
			"APIKeyRequired",
			"Repository",
			"Create",
			"The API key is required",
			http.StatusBadRequest,
		)
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Create(apiKey)
	if result.Error != nil {
		return nil, r.exceptions.NewForDomain(
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
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("last_used_at", usedAt)
	if result.Error != nil {
		return r.exceptions.NewForDomain(
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
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Model(&schemas.APIKey{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", revokedAt)
	if result.Error != nil {
		return r.exceptions.NewForDomain(
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
