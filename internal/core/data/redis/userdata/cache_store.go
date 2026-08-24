package userdata

import (
	"context"
	"strings"

	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	redislibraries "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/libraries"
)

type UserDataCacheStore struct {
	clientSet *platformredis.ClientSet
}

func NewUserDataCacheStore(
	clientSet *platformredis.ClientSet,
) *UserDataCacheStore {
	return &UserDataCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	ctx context.Context,
	clientSet *platformredis.ClientSet,
) (*UserDataCacheStore, error) {
	store := NewUserDataCacheStore(clientSet)
	if err := store.Initialize(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *UserDataCacheStore) Initialize(_ context.Context) error {
	userQuotaLibraryContent := strings.Join([]string{
		"#!lua name=" + redislibraries.UserQuotaLibrary,
		redislibraries.CheckAndUpdateUserQuotaByFormattedKeyContent,
		redislibraries.BestEffortBatchCheckAndUpdateUserQuotasByFormattedKeysContent,
		redislibraries.AllOrNothingBatchCheckAndUpdateUserQuotasByFormattedKeysContent,
		redislibraries.BestEffortBatchCheckAndUpdateUserQuotasByFormattedKeyContent,
		redislibraries.AllOrNothingBatchCheckAndUpdateUserQuotasByFormattedKeyContent,
	}, "\n\n")

	for _, redisClient := range s.clientSet.Clients() {
		if err := redisClient.Do(
			"FUNCTION",
			"LOAD",
			"REPLACE",
			userQuotaLibraryContent,
		).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserDataCacheStore) ClientSet() *platformredis.ClientSet {
	return s.clientSet
}
