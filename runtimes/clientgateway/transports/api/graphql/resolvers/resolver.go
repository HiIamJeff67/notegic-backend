package resolvers

import (
	dataloaders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/graphql/dataloaders"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

// This file is GraphQL transport dependency injection.

type Resolver struct {
	dataloader  dataloaders.Dataloaders
	coreAdapter *coreadapters.CoreAdapter
}

func NewResolver(coreAdapter *coreadapters.CoreAdapter) *Resolver {
	return &Resolver{
		dataloader:  dataloaders.NewDataloaders(coreAdapter),
		coreAdapter: coreAdapter,
	}
}
