package realtimelease

import (
	"context"

	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

type RealtimeLeaseCacheStore struct {
	clientSet *platformredis.ClientSet
}

func NewRealtimeLeaseCacheStore(
	clientSet *platformredis.ClientSet,
) *RealtimeLeaseCacheStore {
	return &RealtimeLeaseCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *platformredis.ClientSet,
) *RealtimeLeaseCacheStore {
	return NewRealtimeLeaseCacheStore(clientSet)
}

func (s *RealtimeLeaseCacheStore) Initialize(_ context.Context) error {
	return nil
}

func (s *RealtimeLeaseCacheStore) ClientSet() *platformredis.ClientSet {
	return s.clientSet
}
