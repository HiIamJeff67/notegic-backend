# Core

Core owns Notegic's durable product rules and workspace data. Other runtimes
call it to act on behalf of a user or to react to a completed product change.

## Feature contribution

| Product area | Core responsibility |
| --- | --- |
| Workspace | Shelves, materials, block packs, sharing, and permissions. |
| Planning | Stations, routines, tasks, dependencies, and task records. |
| Accounts | Authentication, account security, linked identities, and API keys. |
| Collaboration | Access decisions and durable projections for shared documents. |
| Side effects | Publishes reliable requests for notifications, email, realtime updates, and document maintenance. |

## Technical design

Core is a Go application backed by PostgreSQL and Redis. It exposes versioned
internal and GraphQL contracts, keeps domain services and repositories here,
and publishes cross-runtime side effects through its transactional outbox.
Consumers must use the versioned contracts instead of importing Core internals.

For local development, start dependencies with `make compose-up` at the
repository root, then run the database setup below. The command uses the
repository's SOPS-managed environment; do not store decrypted secrets in Git.

## Commands

Run these from this directory:

```sh
make migrate
make seed
make test
make test-race
make test-e2e
```

GraphQL generation is owned by `contracts/`:

```sh
make -C ../../contracts gql-generate
```

Generated artifacts are checked in. Do not hand-edit them.

## Structure

```text
commands/           Runtime and database commands
configs/            Core configuration
data/               PostgreSQL, Redis, and storage integration
services/           Domain services by product area
transports/         Gateway, worker, email, and status boundaries
workers/             Core-owned background work
validations/        Domain input validation
test/               Runtime tests
```

Further reading: [transactional outbox](../../docs/system-design/transactional-outbox.md),
[cross-runtime side effects](../../docs/system-design/cross-runtime-side-effects.md),
and [Core contracts](../../contracts/core/v1/README.md).
