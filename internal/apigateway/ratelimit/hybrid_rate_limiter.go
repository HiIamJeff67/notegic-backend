package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	rate "golang.org/x/time/rate"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/internal/apigateway/configs"
	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/internal/apigateway/data/redis/ratelimitrecord"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/apigateway/data/redis/ratelimitrecord/inputs"
)

type HybridRateLimitTask struct {
	NumOfChangingTokens int32 `json:"numOfChangingTokens"`
	IsAccumulated       bool  `json:"isAccumulated"`
	Retries             int   `json:"retries"`
	MaxRetries          int   `json:"maxRetries"`
}

type HybridRateLimiter struct {
	*rate.Limiter
	LimiterMutex   sync.RWMutex
	UserLimit      int32
	WindowDuration time.Duration

	pendingTasks      map[string]HybridRateLimitTask
	pendingTasksMutex sync.Mutex
	syncInterval      time.Duration
	syncTicker        *time.Ticker
	stopChan          chan struct{}

	BackendServerName   platformredis.BackendServerName
	IsAuthorizedLimiter bool
	cacheClient         *ratelimitrecord.RateLimitRecordCacheClient
}

func NewHybridRateLimiter(config gatewayconfig.RateLimitConfig, isAuthorizedLimiter bool) *HybridRateLimiter {
	syncInterval := config.WindowDuration / time.Duration(config.SynchronizationToWindowDurationRatio)
	syncInterval = max(config.MinSynchronizationInterval, syncInterval)

	hrl := &HybridRateLimiter{
		Limiter:             rate.NewLimiter(config.RateLimit, config.Burst),
		UserLimit:           config.UserLimit,
		WindowDuration:      config.WindowDuration,
		pendingTasks:        make(map[string]HybridRateLimitTask, 0),
		syncInterval:        syncInterval,
		syncTicker:          time.NewTicker(syncInterval),
		stopChan:            make(chan struct{}),
		BackendServerName:   config.BackendServerName,
		IsAuthorizedLimiter: isAuthorizedLimiter,
		cacheClient:         config.CacheClient,
	}

	// initially calling syncLoop() to start syncing periodically
	go hrl.syncLoop()

	return hrl
}

func (hrl *HybridRateLimiter) appendPendingTask(key string, tokens int32) {
	hrl.pendingTasksMutex.Lock()
	defer hrl.pendingTasksMutex.Unlock()

	if existingTask, exists := hrl.pendingTasks[key]; exists {
		hrl.pendingTasks[key] = HybridRateLimitTask{
			NumOfChangingTokens: existingTask.NumOfChangingTokens + tokens,
			IsAccumulated:       true,
			Retries:             existingTask.Retries,
			MaxRetries:          3,
		}
	} else {
		hrl.pendingTasks[key] = HybridRateLimitTask{
			NumOfChangingTokens: tokens,
			IsAccumulated:       true,
			Retries:             0,
			MaxRetries:          3,
		}
	}
}

func (hrl *HybridRateLimiter) reappendPendingTasks(failedTasks map[string]HybridRateLimitTask) {
	hrl.pendingTasksMutex.Lock()
	defer hrl.pendingTasksMutex.Unlock()

	for key, task := range failedTasks {
		if task.Retries < task.MaxRetries {
			hrl.pendingTasks[key] = HybridRateLimitTask{
				NumOfChangingTokens: task.NumOfChangingTokens,
				IsAccumulated:       task.IsAccumulated,
				Retries:             task.Retries + 1,
				MaxRetries:          task.MaxRetries,
			}
		} else {
			logs.NotegicLogger.Warn(context.Background(), fmt.Sprintf("Dropping task for key %s after %d retries", key, task.MaxRetries))
		}
	}
}

func (hrl *HybridRateLimiter) batchSync() {
	hrl.pendingTasksMutex.Lock()
	if len(hrl.pendingTasks) == 0 {
		hrl.pendingTasksMutex.Unlock()
		return
	}

	fetchedPendingTasks := make(map[string]HybridRateLimitTask)
	for key, task := range hrl.pendingTasks {
		fetchedPendingTasks[key] = task
	}
	hrl.pendingTasks = make(map[string]HybridRateLimitTask)
	hrl.pendingTasksMutex.Unlock()

	inputs := make([]cacheinputs.BatchSynchronizeRateLimitRecordCacheInput, 0, len(fetchedPendingTasks))
	for identifier, task := range fetchedPendingTasks {
		inputs = append(inputs, cacheinputs.BatchSynchronizeRateLimitRecordCacheInput{
			Identifier: identifier,
			Input: cacheinputs.SynchronizeRateLimitRecordCacheInput{
				NumOfChangingTokens: task.NumOfChangingTokens,
				IsAccumulated:       task.IsAccumulated,
			},
		})
	}

	if err := hrl.cacheClient.BatchSynchronize(inputs, hrl.BackendServerName); err != nil {
		logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("Failed to batch synchronize rate limits to Redis: %v", err))
		hrl.reappendPendingTasks(fetchedPendingTasks)
	} else if len(inputs) > 0 {
		logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Batch synchronized %d rate limits to Redis", len(inputs)))
	}
}

func (hrl *HybridRateLimiter) syncLoop() {
	for {
		select {
		case <-hrl.syncTicker.C:
			hrl.batchSync()
		case <-hrl.stopChan:
			hrl.batchSync()
			return
		}
	}
}

/* ============================== Operations for Unauthorized Middleware (Client IP based) ============================== */

func (hrl *HybridRateLimiter) checkBucketLimitByFingerprint(fingerprint string, n int32) int32 {
	var totalTokensUsed int32 = 0

	for _, backendServerName := range platformredis.AllBackendServerNames {
		rateLimitRecordCache, exception := hrl.cacheClient.Get(fingerprint, backendServerName)
		if exception != nil {
			continue
		}

		now := time.Now()
		if now.Sub(rateLimitRecordCache.WindowStartTime) >= rateLimitRecordCache.WindowDuration {
			continue
		}

		totalTokensUsed += rateLimitRecordCache.NumOfTokens
	}

	return hrl.UserLimit - totalTokensUsed - n
}

func (hrl *HybridRateLimiter) AllowByFingerprint(fingerprint string) (bool, int32) {
	return hrl.AllowNByFingerprint(fingerprint, time.Now(), 1)
}

func (hrl *HybridRateLimiter) AllowNByFingerprint(fingerprint string, now time.Time, n int) (bool, int32) {
	hrl.LimiterMutex.RLock()
	defer hrl.LimiterMutex.RUnlock()

	// 1. Use the Limiter from the rate utility for fast checking
	if !hrl.Limiter.AllowN(now, n) {
		logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Request blocked by local rate limiter for client IP: %s, requested: %d", fingerprint, n))
		return false, 0
	}

	// 2. Use the rate limit record in redis cache to check if the request from the same source has exceeded some certain count
	remaining := hrl.checkBucketLimitByFingerprint(fingerprint, int32(n))
	if remaining < 0 {
		logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Request blocked by global rate limiter for client IP: %s, requested: %d", fingerprint, n))
		return false, 0
	}

	// 3. Record the tokens that being used, and batch synchronize to the redis
	hrl.appendPendingTask(fingerprint, int32(n))

	return true, remaining
}

/* ============================== Operations for Authorized Middleware (User ID based) ============================== */

func (hrl *HybridRateLimiter) checkBucketLimitByUserId(userId uuid.UUID, n int32) int32 {
	var totalTokensUsed int32 = 0

	for _, backendServerName := range platformredis.AllBackendServerNames {
		rateLimitRecordCache, exception := hrl.cacheClient.Get(userId.String(), backendServerName)
		if exception != nil {
			continue
		}

		now := time.Now()
		if now.Sub(rateLimitRecordCache.WindowStartTime) >= rateLimitRecordCache.WindowDuration {
			continue
		}

		totalTokensUsed += rateLimitRecordCache.NumOfTokens
	}

	return hrl.UserLimit - totalTokensUsed - n
}

func (hrl *HybridRateLimiter) AllowByUserId(userId uuid.UUID) (bool, int32) {
	return hrl.AllowNByUserId(userId, time.Now(), 1)
}

func (hrl *HybridRateLimiter) AllowNByUserId(userId uuid.UUID, now time.Time, n int) (bool, int32) {
	hrl.LimiterMutex.RLock()
	defer hrl.LimiterMutex.RUnlock()

	// 1. Use the Limiter from the rate utility for fast checking
	if !hrl.Limiter.AllowN(now, n) {
		logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Request blocked by local rate limiter for user ID: %s, requested: %d", userId.String(), n))
		return false, 0
	}

	// 2. Use the rate limit record in redis cache to check if the request from the same source has exceeded some certain count
	remaining := hrl.checkBucketLimitByUserId(userId, int32(n))
	if remaining < 0 {
		logs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Request blocked by global rate limiter for user ID: %s, requested: %d", userId.String(), n))
		return false, 0
	}

	// 3. Record the tokens that being used, and batch synchronize to the redis
	hrl.appendPendingTask(userId.String(), int32(n))

	return true, remaining
}

/* ============================== Utility Methods ============================== */

func (hrl *HybridRateLimiter) Allow(key string) (bool, int32) {
	if hrl.IsAuthorizedLimiter {
		userId, err := uuid.Parse(key)
		if err != nil {
			logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("Invalid user ID format: %s", key))
			return false, 0
		}
		return hrl.AllowByUserId(userId)
	} else {
		return hrl.AllowByFingerprint(key)
	}
}

func (hrl *HybridRateLimiter) AllowN(key string, now time.Time, n int) (bool, int32) {
	if hrl.IsAuthorizedLimiter {
		userId, err := uuid.Parse(key)
		if err != nil {
			logs.NotegicLogger.Error(context.Background(), nil, fmt.Sprintf("Invalid user ID format: %s", key))
			return false, 0
		}
		return hrl.AllowNByUserId(userId, now, n)
	} else {
		return hrl.AllowNByFingerprint(key, now, n)
	}
}

func (hrl *HybridRateLimiter) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"userLimit":           hrl.UserLimit,
		"windowDuration":      hrl.WindowDuration,
		"backendServerName":   hrl.BackendServerName,
		"isAuthorizedLimiter": hrl.IsAuthorizedLimiter,
	}
}

func (hrl *HybridRateLimiter) GetDetailStatus() map[string]interface{} {
	hrl.pendingTasksMutex.Lock()
	defer hrl.pendingTasksMutex.Unlock()
	return map[string]interface{}{
		"userLimit":           hrl.UserLimit,
		"windowDuration":      hrl.WindowDuration,
		"isPending":           len(hrl.pendingTasks) > 0,
		"backendServerName":   hrl.BackendServerName,
		"isAuthorizedLimiter": hrl.IsAuthorizedLimiter,
	}
}

func (hrl *HybridRateLimiter) Stop() {
	close(hrl.stopChan)
	hrl.syncTicker.Stop()
}
