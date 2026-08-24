package ratelimitrecord

import (
	"context"

	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	redislibraries "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/ratelimitrecord/libraries"
)

type RateLimitRecordCacheStore struct {
	clientSet *platformredis.ClientSet
}

func NewRateLimitRecordCacheStore(
	clientSet *platformredis.ClientSet,
) *RateLimitRecordCacheStore {
	return &RateLimitRecordCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *platformredis.ClientSet,
) *RateLimitRecordCacheStore {
	return NewRateLimitRecordCacheStore(clientSet)
}

func (s *RateLimitRecordCacheStore) Initialize(_ context.Context) error {
	for _, redisClient := range s.clientSet.Clients() {
		if err := redisClient.Do(
			"FUNCTION",
			"LOAD",
			"REPLACE",
			redislibraries.RateLimitRecordLibraryContent,
		).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s *RateLimitRecordCacheStore) ClientSet() *platformredis.ClientSet {
	return s.clientSet
}
