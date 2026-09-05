# DurableJob

DurableJob performs scheduled, retryable work for routine tasks. It lets a
routine create work at the right time and records the result without tying that
work to an active user request.

## Feature contribution

| Capability | DurableJob responsibility |
| --- | --- |
| Routine scheduling | Plans and claims due routine tasks. |
| Task execution | Resolves task templates, applies work, and records results. |
| Reliable recovery | Recovers stale work and keeps execution state durable. |
| Live progress | Publishes task lifecycle and completion updates for connected clients. |

## Technical design

This Go runtime has a PostgreSQL-backed routine-task engine and local workers.
It uses Core-owned repositories for the product data it needs, calls YjsWorker
when task execution changes a block pack, and publishes lifecycle events for
RealtimeGateway. Its own runtime database access is limited to durable-job
state; business permissions remain with Core.

Start dependencies through `make compose-up` at the repository root. Database
bootstrap and migration are performed as part of the repository's Compose
initialization; do not run against a production database by accident.

## Commands

Run these from this directory:

```sh
make test
make test-race
make test-e2e
```

## Structure

```text
commands/                       Runtime and database commands
configs/                        Worker configuration
data/postgres/                  Durable-job persistence
services/routinetask/           Routine-task planning and execution services
workers/routinetask/            Claiming, scheduling, recovery, and execution engine
transports/core/                Core-facing boundary
transports/yjsworker/           Document update boundary
transports/realtimegateway/     Lifecycle event producer
transports/status/              Health and startup endpoints
test/                           Runtime tests
```

Further reading: [Kafka event contracts](../../docs/system-design/kafka-event-contracts.md)
and [cross-runtime side effects](../../docs/system-design/cross-runtime-side-effects.md).
