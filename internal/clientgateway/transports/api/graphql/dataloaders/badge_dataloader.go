package dataloaders

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	gophersdataloader "github.com/graph-gophers/dataloader/v7"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/badges"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type LoadBadgeSource string

const (
	LoadBadgeSourceUserPublicId LoadBadgeSource = "LoadBadgeSourceUserPublicId"
)

type BadgeLoaderKey struct {
	PublicId uuid.UUID       `json:"publicId"`
	Source   LoadBadgeSource `json:"source"`
}

type BadgeLoaderType = gophersdataloader.Loader[BadgeLoaderKey, *cgqlmodels.PublicBadge]
type BadgeBatchFunctionType = gophersdataloader.BatchFunc[BadgeLoaderKey, *cgqlmodels.PublicBadge]
type BadgeResultType = gophersdataloader.Result[*cgqlmodels.PublicBadge]

type BadgeDataloaderInterface interface {
	GetLoader() *BadgeLoaderType
	LoadByUserPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicBadge, error)
}

type BadgeDataloader struct {
	coreAdapter *coreadapters.CoreAdapter
	loader      *BadgeLoaderType
}

func NewBadgeDataloader(coreAdapter *coreadapters.CoreAdapter) BadgeDataloaderInterface {
	dataloader := &BadgeDataloader{
		coreAdapter: coreAdapter,
	}
	dataloader.loader = gophersdataloader.NewBatchedLoader(
		dataloader.batchFunction(),
		gophersdataloader.WithWait[BadgeLoaderKey, *cgqlmodels.PublicBadge](loaderDelayOfBadge),
	)

	return dataloader
}

/* ============================== Dataloader Methods ============================== */

func (d *BadgeDataloader) GetLoader() *BadgeLoaderType {
	return d.loader
}

func (d *BadgeDataloader) batchFunction() BadgeBatchFunctionType {
	return func(ctx context.Context, keys []BadgeLoaderKey) []*BadgeResultType {
		results := make([]*BadgeResultType, len(keys))
		publicIds := make([]uuid.UUID, 0, len(keys))
		indexesByPublicId := make(map[uuid.UUID][]int, len(keys))

		for index, key := range keys {
			if key.Source != LoadBadgeSourceUserPublicId {
				exception := cexceptions.New(
					"InvalidSource",
					"GraphQL",
					"LoadUserBadges",
					"Badge dataloader source is invalid",
					http.StatusInternalServerError,
					true,
				)
				results[index] = &BadgeResultType{
					Error: exception.Origin(),
				}
				continue
			}

			publicIds = append(publicIds, key.PublicId)
			indexesByPublicId[key.PublicId] = append(indexesByPublicId[key.PublicId], index)
		}
		if len(publicIds) == 0 {
			return results
		}

		ginContext, exception := gatewaycontexts.GetAndConvertContextToGinContext(ctx)
		if exception != nil {
			for _, indexes := range indexesByPublicId {
				for _, index := range indexes {
					results[index] = &BadgeResultType{
						Error: exception.Origin(),
					}
				}
			}
			return results
		}

		requestDto := capi.LoadUserBadgesRequestDto(publicIds)
		response, exception := coreadapters.CallSecurly[
			capi.LoadUserBadgesRequestDto,
			capi.LoadUserBadgesResponseDto,
		](
			ginContext,
			d.coreAdapter,
			&requestDto,
			capi.LoadUserBadgesOperation,
			"/core/v1/badges/graphql/load",
		)
		if exception != nil {
			for _, indexes := range indexesByPublicId {
				for _, index := range indexes {
					results[index] = &BadgeResultType{
						Error: exception.Origin(),
					}
				}
			}
			return results
		}

		for index, model := range response.Data {
			if index >= len(publicIds) {
				break
			}
			for _, resultIndex := range indexesByPublicId[publicIds[index]] {
				results[resultIndex] = &BadgeResultType{
					Data: model,
				}
			}
		}

		return results
	}
}

func (d *BadgeDataloader) LoadByUserPublicId(
	ctx context.Context,
	publicId uuid.UUID,
) (*cgqlmodels.PublicBadge, error) {
	future := d.loader.Load(ctx, BadgeLoaderKey{
		PublicId: publicId,
		Source:   LoadBadgeSourceUserPublicId,
	})

	return future()
}
