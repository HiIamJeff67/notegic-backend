package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	cgenerated "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/generated"

	resolvers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/graphql/resolvers"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

func GraphQLHandler(coreAdapter *coreadapters.CoreAdapter) gin.HandlerFunc {
	resolver := resolvers.NewResolver(coreAdapter)
	server := handler.NewDefaultServer(cgenerated.NewExecutableSchema(cgenerated.Config{
		Resolvers: resolver,
	}))

	return func(ctx *gin.Context) {
		requestContext := context.WithValue(
			ctx.Request.Context(),
			sharedcontexts.ContextFieldName_GinContext,
			ctx,
		)
		server.ServeHTTP(ctx.Writer, ctx.Request.WithContext(requestContext))
	}
}

func PlaygroundHandler() gin.HandlerFunc {
	return gin.WrapH(playground.Handler("GraphQL Playground", "/graphql"))
}
