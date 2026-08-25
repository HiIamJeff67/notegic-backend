package ratelimitrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/data/redis/ratelimitrecord/inputs"
	redislibraries "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/data/redis/ratelimitrecord/libraries"
)

type RateLimitRecordCache struct {
	NumOfTokens     int32         `json:"numOfTokens"`
	WindowStartTime time.Time     `json:"windowStartTime"`
	WindowDuration  time.Duration `json:"windowDuration"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type RateLimitRecordCacheClient struct {
	cacheStore *RateLimitRecordCacheStore

	jitterMaxOffset                    time.Duration
	batchSynchronizeFunctionArgvPerKey int
}

/* ============================== Constructor ============================== */

func NewRateLimitRecordCacheClient(cacheStore *RateLimitRecordCacheStore) *RateLimitRecordCacheClient {
	return &RateLimitRecordCacheClient{
		cacheStore: cacheStore,

		jitterMaxOffset:                    5 * time.Second,
		batchSynchronizeFunctionArgvPerKey: 2,
	}
}

/* ============================== Auxiliary Methods ============================== */

func (s *RateLimitRecordCacheClient) getRedisClient(backendServerName platformredis.BackendServerName) (*redis.Client, int, *cexceptions.Exception) {
	if s == nil || s.cacheStore == nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"Rate limit cache client is unavailable",
			http.StatusInternalServerError,
			true,
		)
	}

	redisClient, shardIndex, err := s.cacheStore.ClientSet().ClientForKey(string(backendServerName))
	if err != nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"Rate limit cache client is unavailable",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return redisClient, shardIndex, nil
}

func (s *RateLimitRecordCacheClient) formatRateLimitRecordKey(backendServerName platformredis.BackendServerName, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", platformredis.CachePurpose_RateLimit.String(), backendServerName, identifier)
}

func (s *RateLimitRecordCacheClient) calculateExpiration(identifier string, windowStart time.Time, windowDuration time.Duration) time.Duration {
	baseExpiration := windowStart.Add(windowDuration).Sub(time.Now())
	if baseExpiration < 0 {
		return 1
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(identifier))
	random := rand.New(rand.NewSource(int64(hash.Sum32())))

	return baseExpiration + time.Duration(random.Int63n(int64(s.jitterMaxOffset)))
}

/* ============================== CRUD Method ============================== */

func (s *RateLimitRecordCacheClient) Get(
	identifier string,
	backendServerName platformredis.BackendServerName,
) (*RateLimitRecordCache, *cexceptions.Exception) {
	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return nil, exception
	}

	cacheString, err := redisClient.Get(s.formatRateLimitRecordKey(backendServerName, identifier)).Result()
	if err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"Cache",
			"GetRateLimitRecord",
			"Rate limit record was not found",
			http.StatusNotFound,
			true,
		).WithOrigin(err)
	}

	var rateLimitRecordCache RateLimitRecordCache
	if err := json.Unmarshal([]byte(cacheString), &rateLimitRecordCache); err != nil {
		return nil, cexceptions.New(
			"DeserializationFailed",
			"Cache",
			"GetRateLimitRecord",
			"Failed to decode the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully got cached rate limit record from Redis shard %d", shardIndex))
	return &rateLimitRecordCache, nil
}

func (s *RateLimitRecordCacheClient) Set(
	identifier string,
	backendServerName platformredis.BackendServerName,
	rateLimitRecordCache RateLimitRecordCache,
) *cexceptions.Exception {
	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return exception
	}

	value, err := json.Marshal(rateLimitRecordCache)
	if err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"SetRateLimitRecord",
			"Failed to encode the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	expiresIn := s.calculateExpiration(identifier, rateLimitRecordCache.WindowStartTime, rateLimitRecordCache.WindowDuration)
	if err := redisClient.Set(s.formatRateLimitRecordKey(backendServerName, identifier), string(value), expiresIn).Err(); err != nil {
		return cexceptions.New(
			"FailedToCreate",
			"Cache",
			"SetRateLimitRecord",
			"Failed to store the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully set cached rate limit record in Redis shard %d", shardIndex))
	return nil
}

func (s *RateLimitRecordCacheClient) Update(
	identifier string,
	backendServerName platformredis.BackendServerName,
	input cacheinputs.SynchronizeRateLimitRecordCacheInput,
) *cexceptions.Exception {
	rateLimitRecordCache, exception := s.Get(identifier, backendServerName)
	if exception != nil {
		return exception
	}

	if (!input.IsAccumulated && rateLimitRecordCache.NumOfTokens < input.NumOfChangingTokens) || rateLimitRecordCache.NumOfTokens < 0 {
		return cexceptions.New(
			"InvalidRateLimitTokenCount",
			"RateLimit",
			"UpdateRateLimitRecord",
			"Rate limit token count is invalid",
			http.StatusInternalServerError,
			true,
		)
	}

	if input.IsAccumulated {
		rateLimitRecordCache.NumOfTokens += input.NumOfChangingTokens
	} else {
		rateLimitRecordCache.NumOfTokens -= input.NumOfChangingTokens
	}
	rateLimitRecordCache.UpdatedAt = time.Now()

	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return exception
	}

	value, err := json.Marshal(rateLimitRecordCache)
	if err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"UpdateRateLimitRecord",
			"Failed to encode the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	expiresIn := s.calculateExpiration(identifier, rateLimitRecordCache.WindowStartTime, rateLimitRecordCache.WindowDuration)
	if err := redisClient.Set(s.formatRateLimitRecordKey(backendServerName, identifier), string(value), expiresIn).Err(); err != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"Cache",
			"UpdateRateLimitRecord",
			"Failed to update the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully updated cached rate limit record in Redis shard %d", shardIndex))
	return nil
}

func (s *RateLimitRecordCacheClient) Delete(
	identifier string,
	backendServerName platformredis.BackendServerName,
) *cexceptions.Exception {
	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return exception
	}

	if err := redisClient.Del(s.formatRateLimitRecordKey(backendServerName, identifier)).Err(); err != nil {
		return cexceptions.New(
			"FailedToDelete",
			"Cache",
			"DeleteRateLimitRecord",
			"Failed to delete the rate limit record",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully deleted cached rate limit record from Redis shard %d", shardIndex))
	return nil
}

/* ============================== Batch Method ============================== */

func (s *RateLimitRecordCacheClient) BatchSynchronize(
	inputs []cacheinputs.BatchSynchronizeRateLimitRecordCacheInput,
	backendServerName platformredis.BackendServerName,
) *cexceptions.Exception {
	if len(inputs) == 0 {
		return nil
	}

	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return exception
	}

	keys := make([]interface{}, 0, len(inputs))
	arguments := make([]interface{}, 0, len(inputs)*s.batchSynchronizeFunctionArgvPerKey)
	for _, input := range inputs {
		keys = append(keys, s.formatRateLimitRecordKey(backendServerName, input.Identifier))
		arguments = append(arguments, input.Input.NumOfChangingTokens, input.Input.IsAccumulated)
	}

	command := []interface{}{
		"FCALL",
		redislibraries.BatchSynchronizeRateLimitRecordByFormattedKeysFunction,
		len(keys),
	}
	command = append(command, keys...)
	command = append(command, arguments...)
	if _, err := redisClient.Do(command...).Result(); err != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"Cache",
			"BatchSynchronizeRateLimitRecords",
			"Failed to synchronize rate limit records",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully batch synchronized rate limit records in Redis shard %d", shardIndex))
	return nil
}

func (s *RateLimitRecordCacheClient) BatchDelete(
	identifiers []string,
	backendServerName platformredis.BackendServerName,
) *cexceptions.Exception {
	if len(identifiers) == 0 {
		return nil
	}

	redisClient, shardIndex, exception := s.getRedisClient(backendServerName)
	if exception != nil {
		return exception
	}

	keys := make([]interface{}, 0, len(identifiers))
	for _, identifier := range identifiers {
		keys = append(keys, s.formatRateLimitRecordKey(backendServerName, identifier))
	}

	command := []interface{}{
		"FCALL",
		redislibraries.BatchDeleteRateLimitRecordByFormattedKeysFunction,
		len(keys),
	}
	command = append(command, keys...)
	if _, err := redisClient.Do(command...).Result(); err != nil {
		return cexceptions.New(
			"FailedToDelete",
			"Cache",
			"BatchDeleteRateLimitRecords",
			"Failed to delete rate limit records",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully batch deleted rate limit records from Redis shard %d", shardIndex))
	return nil
}
