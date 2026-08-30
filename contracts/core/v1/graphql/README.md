# Core GraphQL contracts

Core owns this GraphQL source; Gateway owns only its execution, resolvers, and
dataloaders. This directory contains source shared with client code generation:

- `schemas/` contains SDL, including scalar and enum definitions.
- `fragments/` and `queries/` contain reusable client operation source.
- GraphQL RequestDto/ResponseDto are owned by the matching
  `contracts/core/v1/api/<route-domain>/search.go` package.

The Go server generated artifacts, generated models, and scalar implementation
belong to `contracts/core/v1/graphql/generated/`, `contracts/core/v1/graphql/models/`, and
`contracts/core/v1/graphql/scalars/`. Gateway GraphQL execution, resolvers, and
dataloaders belong to `runtimes/clientgateway/transports/api/graphql/`; they access Core
only through `runtimes/clientgateway/transports/core/adapters/CoreAdapter`. Each Core
domain endpoint/router owns its own GraphQL operation; a shared GraphQLEndpoint
or central Core GraphQL router is not permitted.

Run `make -C contracts gql-generate` after changing SDL. Generated output must not be edited
by hand.
