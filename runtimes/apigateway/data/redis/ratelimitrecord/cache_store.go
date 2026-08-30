package ratelimitrecord

import (
	"context"

	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	redislibraries "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/data/redis/ratelimitrecord/libraries"
)

type RateLimitRecordCacheStore struct {
	clientSet *sredis.ClientSet
}

func NewRateLimitRecordCacheStore(
	clientSet *sredis.ClientSet,
) *RateLimitRecordCacheStore {
	return &RateLimitRecordCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *sredis.ClientSet,
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

func (s *RateLimitRecordCacheStore) ClientSet() *sredis.ClientSet {
	return s.clientSet
}
