package apikey

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

const defaultCacheExpiresIn = 5 * time.Minute

type APIKeyCache struct {
	Id           uuid.UUID  `json:"id"`
	UserId       uuid.UUID  `json:"userId"`
	UserPublicId uuid.UUID  `json:"userPublicId"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	RevokedAt    *time.Time `json:"revokedAt"`
}

type APIKeyCacheClient struct {
	cacheStore *APIKeyCacheStore
	expiresIn  time.Duration
}

/* ============================== Constructor ============================== */

func NewAPIKeyCacheClient(cacheStore *APIKeyCacheStore) *APIKeyCacheClient {
	return &APIKeyCacheClient{
		cacheStore: cacheStore,
		expiresIn:  defaultCacheExpiresIn,
	}
}

/* ============================== Auxiliary Methods ============================== */

func (s *APIKeyCacheClient) getRedisClient(identifier string) (*redis.Client, int, *cexceptions.Exception) {
	if s == nil || s.cacheStore == nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"API key cache client is unavailable",
			http.StatusInternalServerError,
			true,
		)
	}

	redisClient, shardIndex, err := s.cacheStore.ClientSet().ClientForKey(identifier)
	if err != nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"API key cache client is unavailable",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return redisClient, shardIndex, nil
}

func (s *APIKeyCacheClient) formatAPIKeyCacheKey(keyHash string) string {
	return fmt.Sprintf("%s:%s", platformredis.CachePurpose_APIKey.String(), keyHash)
}

/* ============================== CRUD Method ============================== */

func (s *APIKeyCacheClient) Get(keyHash string) (*APIKeyCache, *cexceptions.Exception) {
	redisClient, shardIndex, exception := s.getRedisClient(keyHash)
	if exception != nil {
		return nil, exception
	}

	cacheString, err := redisClient.Get(s.formatAPIKeyCacheKey(keyHash)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"Cache",
			"GetAPIKey",
			"Cached API key was not found",
			http.StatusNotFound,
			true,
		).WithOrigin(err)
	}

	var apiKeyCache APIKeyCache
	if err := json.Unmarshal([]byte(cacheString), &apiKeyCache); err != nil {
		return nil, cexceptions.New(
			"DeserializationFailed",
			"Cache",
			"GetAPIKey",
			"Failed to decode cached API key",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully got cached API key from Redis shard %d", shardIndex))
	return &apiKeyCache, nil
}

func (s *APIKeyCacheClient) Set(keyHash string, apiKeyCache APIKeyCache) *cexceptions.Exception {
	redisClient, shardIndex, exception := s.getRedisClient(keyHash)
	if exception != nil {
		return exception
	}

	value, err := json.Marshal(apiKeyCache)
	if err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"SetAPIKey",
			"Failed to encode cached API key",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := redisClient.Set(s.formatAPIKeyCacheKey(keyHash), string(value), s.expiresIn).Err(); err != nil {
		return cexceptions.New(
			"FailedToCreate",
			"Cache",
			"SetAPIKey",
			"Failed to store cached API key",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully set cached API key in Redis shard %d", shardIndex))
	return nil
}

func (s *APIKeyCacheClient) Delete(keyHash string) *cexceptions.Exception {
	redisClient, shardIndex, exception := s.getRedisClient(keyHash)
	if exception != nil {
		return exception
	}

	if err := redisClient.Del(s.formatAPIKeyCacheKey(keyHash)).Err(); err != nil {
		return cexceptions.New(
			"FailedToDelete",
			"Cache",
			"DeleteAPIKey",
			"Failed to delete cached API key",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully deleted cached API key from Redis shard %d", shardIndex))
	return nil
}
