package dataloaders

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	gophersdataloader "github.com/graph-gophers/dataloader/v7"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type LoadUserSource string

const (
	LoadUserSourceThemePublicId LoadUserSource = "LoadUserSourceThemePublicId"
)

type UserLoaderKey struct {
	PublicId uuid.UUID      `json:"publicId"`
	Source   LoadUserSource `json:"source"`
}

type UserLoaderType = gophersdataloader.Loader[UserLoaderKey, *cgqlmodels.PublicUser]
type UserBatchFunctionType = gophersdataloader.BatchFunc[UserLoaderKey, *cgqlmodels.PublicUser]
type UserResultType = gophersdataloader.Result[*cgqlmodels.PublicUser]

type UserDataloaderInterface interface {
	GetLoader() *UserLoaderType
	LoadByThemePublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicUser, error)
}

type UserDataloader struct {
	coreAdapter *coreadapters.CoreAdapter
	loader      *UserLoaderType
}

func NewUserDataloader(coreAdapter *coreadapters.CoreAdapter) UserDataloaderInterface {
	dataloader := &UserDataloader{
		coreAdapter: coreAdapter,
	}
	dataloader.loader = gophersdataloader.NewBatchedLoader(
		dataloader.batchFunction(),
		gophersdataloader.WithWait[UserLoaderKey, *cgqlmodels.PublicUser](loaderDelayOfUser),
	)

	return dataloader
}

/* ============================== Dataloader Methods ============================== */

func (d *UserDataloader) GetLoader() *UserLoaderType {
	return d.loader
}

func (d *UserDataloader) batchFunction() UserBatchFunctionType {
	return func(ctx context.Context, keys []UserLoaderKey) []*UserResultType {
		results := make([]*UserResultType, len(keys))
		publicIds := make([]uuid.UUID, 0, len(keys))
		indexesByPublicId := make(map[uuid.UUID][]int, len(keys))

		for index, key := range keys {
			if key.Source != LoadUserSourceThemePublicId {
				exception := cexceptions.New(
					"InvalidSource",
					"GraphQL",
					"LoadThemeAuthors",
					"User dataloader source is invalid",
					http.StatusInternalServerError,
					true,
				)
				results[index] = &UserResultType{
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
					results[index] = &UserResultType{
						Error: exception.Origin(),
					}
				}
			}
			return results
		}

		requestDto := capi.LoadThemeAuthorsRequestDto(publicIds)
		response, exception := coreadapters.CallSecurly[
			capi.LoadThemeAuthorsRequestDto,
			capi.LoadThemeAuthorsResponseDto,
		](
			ginContext,
			d.coreAdapter,
			&requestDto,
			capi.LoadThemeAuthorsOperation,
			"/core/v1/users/graphql/load-theme-authors",
		)
		if exception != nil {
			for _, indexes := range indexesByPublicId {
				for _, index := range indexes {
					results[index] = &UserResultType{
						Error: exception.Origin(),
					}
				}
			}
			return results
		}

		for index, publicUser := range response.Data {
			if index >= len(publicIds) {
				break
			}
			for _, resultIndex := range indexesByPublicId[publicIds[index]] {
				results[resultIndex] = &UserResultType{
					Data: publicUser,
				}
			}
		}

		return results
	}
}

func (d *UserDataloader) LoadByThemePublicId(
	ctx context.Context,
	publicId uuid.UUID,
) (*cgqlmodels.PublicUser, error) {
	future := d.loader.Load(ctx, UserLoaderKey{
		PublicId: publicId,
		Source:   LoadUserSourceThemePublicId,
	})

	return future()
}
