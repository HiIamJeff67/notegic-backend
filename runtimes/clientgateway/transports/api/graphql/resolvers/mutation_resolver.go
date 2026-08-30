package resolvers

type MutationResolverInterface interface{}

type MutationResolver struct{ *Resolver }

func (r *Resolver) Mutation() *MutationResolver {
	return &MutationResolver{r}
}
