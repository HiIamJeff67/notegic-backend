package realtimelease

import (
	"context"

	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

type RealtimeLeaseCacheStore struct {
	clientSet *sredis.ClientSet
}

func NewRealtimeLeaseCacheStore(
	clientSet *sredis.ClientSet,
) *RealtimeLeaseCacheStore {
	return &RealtimeLeaseCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *sredis.ClientSet,
) *RealtimeLeaseCacheStore {
	return NewRealtimeLeaseCacheStore(clientSet)
}

func (s *RealtimeLeaseCacheStore) Initialize(_ context.Context) error {
	return nil
}

func (s *RealtimeLeaseCacheStore) ClientSet() *sredis.ClientSet {
	return s.clientSet
}
