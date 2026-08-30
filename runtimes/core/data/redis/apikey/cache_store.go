package apikey

import (
	"context"

	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

type APIKeyCacheStore struct {
	clientSet *sredis.ClientSet
}

func NewAPIKeyCacheStore(
	clientSet *sredis.ClientSet,
) *APIKeyCacheStore {
	return &APIKeyCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *sredis.ClientSet,
) *APIKeyCacheStore {
	return NewAPIKeyCacheStore(clientSet)
}

func (s *APIKeyCacheStore) Initialize(_ context.Context) error {
	return nil
}

func (s *APIKeyCacheStore) ClientSet() *sredis.ClientSet {
	if s == nil {
		return nil
	}
	return s.clientSet
}
