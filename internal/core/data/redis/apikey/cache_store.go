package apikey

import (
	"context"

	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

type APIKeyCacheStore struct {
	clientSet *platformredis.ClientSet
}

func NewAPIKeyCacheStore(
	clientSet *platformredis.ClientSet,
) *APIKeyCacheStore {
	return &APIKeyCacheStore{
		clientSet: clientSet,
	}
}

func Register(
	_ context.Context,
	clientSet *platformredis.ClientSet,
) *APIKeyCacheStore {
	return NewAPIKeyCacheStore(clientSet)
}

func (s *APIKeyCacheStore) Initialize(_ context.Context) error {
	return nil
}

func (s *APIKeyCacheStore) ClientSet() *platformredis.ClientSet {
	if s == nil {
		return nil
	}
	return s.clientSet
}
