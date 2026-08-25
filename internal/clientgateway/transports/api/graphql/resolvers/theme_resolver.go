package resolvers

import (
	"context"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
)

type ThemeResolverInterface interface{}

type ThemeResolver struct {
	*Resolver
}

func NewThemeResolver(resolver *Resolver) ThemeResolverInterface {
	return &ThemeResolver{
		Resolver: resolver,
	}
}

/* ============================== Resolver Methods ============================== */
// [MainSchema(as the filename) ---Indicator of MainSchema---> RelativeSchema(has the relationship between the MainSchema)]

// [PublicTheme ---PublicTheme.PublicId---> PublicUser]
func (r *ThemeResolver) Author(ctx context.Context, obj *cgqlmodels.PublicTheme) (*cgqlmodels.PublicUser, error) {
	return r.dataloader.UserDataLoader.LoadByThemePublicId(ctx, obj.PublicID)
}
