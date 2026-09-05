# YjsWorker

YjsWorker owns the durable shared-document side of Notegic block packs. It
allows collaborators to work on the same document while keeping a durable
document history and a queryable projection for the rest of the product.

## Feature contribution

| Capability | YjsWorker responsibility |
| --- | --- |
| Shared editing | Applies and persists Yjs document updates for block packs. |
| Durable documents | Restores document state and creates snapshots when needed. |
| Product visibility | Projects document content for Core's product-facing reads. |
| Maintenance | Compacts document history after Core sends maintenance hints. |

## Technical design

This TypeScript runtime uses Yjs, BlockNote-compatible document structures,
PostgreSQL persistence, Kafka commands and replies, and HTTP endpoints for
runtime-to-runtime work. RealtimeGateway transports live updates; Core owns
permissions and product metadata. Keep raw Yjs updates and snapshots inside
this boundary rather than adding them to general event contracts.

The full development stack starts from the repository root with `make compose-up`.
Use `pnpm` in this runtime, and run integration tests only against the intended
test database.

## Commands

Run these from this directory:

```sh
pnpm run dev
pnpm run build
pnpm test
pnpm run format:check
pnpm run test:integration
```

`test:integration` requires the integration PostgreSQL service. The repository
root target `make test-integration-yjs` supplies its expected connection values.

## Structure

```text
configs/             Runtime configuration
data/postgres/       Yjs persistence and projection storage
services/            Document, projection, and compaction logic
transports/core/     Core command and reply boundary
transports/realtime/ RealtimeGateway boundary
transports/status/   Health and startup endpoints
types/               Runtime-local TypeScript types
util/                Runtime helpers
main.ts              Application entry point
```

Further reading: [Yjs collaboration design](../../docs/system-design/yjs-collaboration.md),
[Kafka event contracts](../../docs/system-design/kafka-event-contracts.md), and
[YjsWorker contracts](../../contracts/yjs-worker/v1/README.md).
