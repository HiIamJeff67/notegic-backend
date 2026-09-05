# RealtimeGateway

RealtimeGateway provides live connections for Notegic applications. It keeps
clients informed about document collaboration, presence, notification, and
routine-task changes without making it the owner of durable product data.

## Feature contribution

| Capability | RealtimeGateway responsibility |
| --- | --- |
| Live collaboration | Admits permitted clients to block-pack document channels. |
| Presence and sessions | Tracks active connections and applies session revocation. |
| In-app updates | Delivers notifications and routine-task lifecycle changes to connected clients. |
| Connection safety | Verifies tickets, enforces admission limits, and manages connection leases. |

## Technical design

This Go runtime serves HTTP and WebSocket traffic with Gin and Gorilla
WebSocket. Redis holds realtime connection, lease, and rate-limit state; Kafka
consumers receive lifecycle events from Core, DurableJob, and Notification.
YjsWorker owns document persistence, so this runtime only carries live document
traffic and must not become a second persistence layer.

Start the complete local stack from the repository root with `make compose-up`.
Its environment is encrypted with SOPS. Read the protocol and event-boundary
documents before changing frames or delivery semantics.

## Commands

Run these from this directory:

```sh
make test
make test-race
make test-e2e
```

## Structure

```text
commands/                Runtime command entry point
configs/                 Gateway configuration
data/redis/              Connection, lease, and rate-limit state
ratelimit/               Admission and request-rate policy
transports/api/          Ticket and HTTP endpoints
transports/websocket/    WebSocket protocol and connection handling
transports/core/         Core lifecycle consumers
transports/durablejob/   Routine-task consumers
transports/notification/ Notification consumers
transports/status/       Health and startup endpoints
workers/                 Local realtime workers
test/                    Runtime tests
```

Further reading: [realtime protocol](../../docs/system-design/realtime-protocol.md),
[Kafka event contracts](../../docs/system-design/kafka-event-contracts.md), and
[realtime contracts](../../contracts/realtime-gateway/v1/README.md).
