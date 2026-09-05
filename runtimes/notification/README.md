# Notification

Notification owns a user's persistent in-app notification inbox. It records
notices created by product activity and makes them available to Notegic clients
through the authenticated gateway path.

## Feature contribution

| Capability | Notification responsibility |
| --- | --- |
| In-app inbox | Stores, searches, marks read, and removes user notifications. |
| Timely updates | Publishes newly created notifications for live delivery. |
| Safe lifecycle | Deduplicates incoming requests, removes expired records, and prevents deleted users from receiving recreated records. |

## Technical design

This Go runtime owns a separate PostgreSQL database. It consumes Core
notification requests from Kafka, processes them idempotently, and writes its
own outbox event for RealtimeGateway. ClientGateway is the public
authentication boundary and calls this runtime through its internal adapter;
do not expose Notification's internal endpoints directly to clients.

Start the complete local environment with `make compose-up` from the repository
root. Its PostgreSQL and Kafka settings come from the SOPS-managed environment;
keep local credentials out of Git.

## Commands

Run these from this directory:

```sh
make test
make test-race
go run ./commands migrateDB
```

The Compose initialization normally applies the runtime migration, so use the
direct migration command only when deliberately operating the configured local
database.

## Structure

```text
commands/             Runtime and database commands
configs/              PostgreSQL, consumer, and cleanup configuration
data/postgres/        Notification persistence
services/             Inbox and notification lifecycle services
transports/core/      Core request consumer
transports/gateway/   Internal gateway endpoints
validations/          Notification payload validation
workers/              Cleanup and retention work
test/                 Runtime tests
```

Further reading: [Notification runtime design](../../docs/codebase-design/notification-runtime.md)
and [Kafka event contracts](../../docs/system-design/kafka-event-contracts.md).
