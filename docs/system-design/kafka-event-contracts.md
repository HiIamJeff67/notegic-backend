# Kafka Event Contracts

## Scope

This document defines the versioned Kafka contracts used across service
boundaries. The runtime-neutral envelope is owned by
[`contracts/types/events/`](../../contracts/types/events/); Core lifecycle payloads remain in
[`contracts/core/v1/events`](../../contracts/core/v1/events/), while
DurableJob and YjsWorker interaction contracts are owned by their respective
runtime contract packages. Core writes its lifecycle events through the
transactional outbox and relay. RealtimeGateway consumes them, fans them out
through its runtime-owned Redis Pub/Sub channel, and executes local
socket/channel cleanup. NOT-33 supplies consumer idempotency, retry, and DLQ
behavior.

Events state a fact already committed by the producer and a policy already
decided by Core. Consumers execute that fact or policy; they never derive
membership, quota, hierarchy, or room admission rules from local state.

## Envelope

Every message is a JSON `eventscontract.EventEnvelope[D]` with these fields.

| Field | Meaning |
| --- | --- |
| `schemaVersion` | Contract major version, initially `v1`. |
| `eventId` | Globally unique UUID used by consumer idempotency. |
| `eventType` | Stable lifecycle fact name. |
| `aggregateType`, `aggregateId` | Owner of ordering and causal identity. |
| `kafkaKey` | Exact `aggregateId` UUID string. |
| `occurredAt` | UTC time at which Core committed the fact. |
| `correlationId` | Required workflow/request identity propagated across related work. |
| `causationId` | Optional parent event UUID. |
| `trace` | Optional W3C `traceParent` and `traceState`; no browser credentials. |
| `data` | Event-type-specific minimal payload. |

`User` aggregate IDs and `TargetUserPublicId` values are user public UUIDs.
Core internal user IDs must never cross this boundary.

## Initial lifecycle events

| Event type | Aggregate | Data | RealtimeGateway action |
| --- | --- | --- |
| `BlockPackAccessRevoked` | one `BlockPack` | optional `targetUserPublicId` | Detach the target user's channel, or every channel when absent. |
| `BlockPackRoomPolicyChanged` | one `BlockPack` | policy version, strategy, maximum subscribers | Apply the attested policy to later admissions; the current v1 strategy rejects new subscribers. |
| `RootShelfPermissionRevoked` | one `RootShelf` | optional `targetUserPublicId` | Lifecycle/audit fact. Core also emits one BlockPack revoke per affected descendant for channel detachment. |
| `RootShelfPermissionChanged` | one `RootShelf` | `resourceId`, `targetUserPublicId`, permission | User-scoped cache/sidebar invalidation; clients refetch through the API. |
| `RootShelfDeleted` | one `RootShelf` | `resourceId`, `targetUserPublicId` | User-scoped resource invalidation after the soft delete commits. |
| `BlockPackChanged` | one `BlockPack` | `resourceId` | Notify active subscribers of metadata changes. |
| `BlockPackDeleted` | one `BlockPack` | `resourceId` | Notify active subscribers after deletion; channel revocation is emitted alongside it. |
| `UserSessionsRevoked` | one `User` public UUID | empty | Fan out a local session revoke and close the user's active RealtimeGateway connections. |

`SubShelf` is an initial aggregate type for future lifecycle facts. It has no
separate v1 event yet: mutations that must detach active rooms emit the affected
one-BlockPack `BlockPackAccessRevoked` events.

## Topic and ordering strategy

The initial topic is `notegic.core.lifecycle.v1`.

| Property | Rule |
| --- | --- |
| Producer | Core only. |
| Consumer group | A runtime-specific group, for example `notegic-realtimegateway-lifecycle-v1`. |
| Partition key | `aggregateId.String()` exactly; it is copied into `kafkaKey`. |
| Ordering | Kafka ordering is guaranteed only for the same aggregate key/partition. No cross-aggregate ordering is assumed. |
| Partitions | Deploy with 12 production partitions initially; increase only by adding partitions and never rely on cross-key order. |
| Retention | Seven days, delete-based cleanup; this topic is a transport log, not the event-sourced source of truth. |
| Replication | Production uses replication factor three and `min.insync.replicas=2`. |

A bulk lifecycle mutation performs a batch outbox insert. It emits one row and
one Kafka record for each affected BlockPack, keyed by that BlockPack ID; it
never emits a list of unrelated BlockPack IDs in one record. This preserves the
only ordering guarantee consumers may rely on.

## Compatibility and payload limits

Within a major contract version, producers may add optional JSON fields only.
Consumers must ignore unknown fields. Renaming, changing a field type or
meaning, making an optional field required, or changing an event's ordering
meaning requires a versioned `contracts/types/events/` contract and a new `.v2` topic. Producers publish
the new topic independently until all consumers migrate.

Lifecycle payloads contain identifiers and already-decided policy values only.
They must not contain complete entity snapshots, names/profile fields, database
internal IDs, browser cookies, access/refresh/delegation tokens, Yjs updates,
awareness, presence, Redis leases, or WebSocket frames.

## YjsWorker command/reply

Yjs persistence is a separate, versioned Kafka boundary in
[`contracts/yjs-worker/v1`](../../contracts/yjs-worker/v1/). YjsWorker produces
an `EventEnvelope` carrying a `CommandEnvelope` to
`notegic.yjsworker.core.command.v1`, keyed by `blockPackId`. Core consumes it
with the `notegic-core-yjsworker-v1` group and writes a matching reply envelope
to its transactional outbox. The outbox relay publishes the reply to
`notegic.core.yjsworker.reply.v1`, again keyed by the same BlockPack ID.

The command family is `LoadYjsDocument`, `AppendYjsUpdate`,
`LoadCompactableYjsDocument`, `ApplyCompactedYjsDocument`, and
`ApplyBlockProjection`. Each carries its own command UUID, correlation ID,
optional causation ID, trace metadata, producer identity, and BlockPack
partition key. Core echoes those values in its reply. Append deduplicates by
`persistenceBatchId`; compaction and projection use their durable sequence
checkpoints, so Kafka redelivery cannot apply a newer state twice.

Every Core mutation and its reply outbox row share one database transaction.
That prevents a committed mutation from losing its reply if Core stops before
Kafka is available. The YjsWorker editor path waits only for the producer's
replicated Kafka ACK; the Core application reply is consumed asynchronously as
the authoritative persistence watermark. A rejected or timed-out application
reply causes the affected room to resync, while a retried command reuses its
persistence batch ID. Invalid envelopes are routed to the normal command DLQ.
RealtimeGateway does not relay Yjs persistence or projection HTTP requests.

## Core Yjs maintenance coordination

Core owns the maintenance worker. DurableJob does not consume Yjs maintenance
hints and does not exchange Yjs maintenance requests or results with Core.
Core's worker consumes the Core-owned hint, reads the shared PostgreSQL state,
and sends the minimal command directly to YjsWorker.

Core writes `YjsMaintenanceHint` to
`notegic.core.yjs-maintenance-hint.v1` in the same transaction as a
new BlockPack Yjs document or accepted update. Core's worker coalesces and
prioritizes those hints by BlockPack partition key, then publishes a compact or
project request to `notegic.core.yjsworker.maintenance-command.v1`.

The worker reads and computes against Core through the existing command/reply
boundary; it does not receive a snapshot in the maintenance event. The result
is published on `notegic.yjsworker.core.maintenance-result.v1` and consumed by
Core's worker for bounded retry. The two Core/YjsWorker topics use the BlockPack
UUID as the key and have local DLQ counterparts.

Maintenance events contain only IDs, watermarks, sizes, operation and status.
They never carry raw Yjs updates, snapshots, state vectors, or BlockNote block
trees. A missing result is recoverable through the next Core hint and the
reconciliation process; no Core transaction waits for DurableJob or the worker.
Core's runtime-owned
`internal/core/workers/yjs_maintenance_reconciliation_worker.go` runs
an immediate startup scan and a low-frequency hourly reconciliation scan for
documents whose durable sequence is still ahead of a compaction or projection
watermark. The scan emits only fresh metadata hints through the same
transactional outbox; it never reads or sends document bytes to DurableJob.

## Consumer reliability and idempotency

Every runtime consumer is built on `shared/platform/kafka.Consumer`. It uses
manual offset commits, bounded `PollRecords`, and Kafka's
`BlockRebalanceOnPoll`: an offset is committed only after its handler has
completed successfully, or after the original record has been durably written
to its dead-letter topic. The consumer always releases the rebalance gate when
the batch ends. Transient broker dial, DNS, and group-session poll failures also
release the gate and retry with the configured bounded exponential backoff
instead of terminating the runtime consumer. A process crash, handler panic,
failed dead-letter publish, or failed offset commit therefore leaves the source
record uncommitted for redelivery.

## Transport ownership

Kafka integration is runtime-specific transport code, not a generic worker
layer. Core's inbound consumers are grouped by the runtime they receive from:

```text
internal/core/transports/
  durablejob/consumers/
  durablejob/eventbuilders/
  durablejob/producers/
  yjsworker/consumers/
  yjsworker/producers/
  outbox_relay.go
```

DurableJob has no Core-facing Kafka transport:

```text
internal/durablejob/transports/core/
  (empty)
```

The routine-task engine and executor do not construct Kafka clients or encode
Core event envelopes. DurableJob writes execution results and business
mutations through its own PostgreSQL connection. Kafka remains only for the
RealtimeGateway lifecycle hint.

Core transport code distinguishes two event-construction roles:

* `eventbuilders/` contains `*_event_builder.go` files with `Build()` methods.
  They create envelopes for events that must first be inserted into Core's
  transactional outbox.
* `producers/` contains `*_producer.go` files with `Produce()` methods. They
  marshal and publish events immediately through the platform Kafka producer
  when no Core database transaction needs to be atomically coupled to the
  publish.

An event builder does not publish a Kafka record by itself. `OutboxRelay`
claims its persisted envelope and is the component that calls the platform
producer. This distinction is intentional and must remain visible in the
folder and file names.

### Core/DurableJob boundary

There is no Core/DurableJob Kafka producer/consumer pair for routine tasks or
Yjs maintenance:

| Concern | Owner | Boundary |
| --- | --- | --- |
| RoutineTask claim, quota, execution, finalization | DurableJob | Shared PostgreSQL, DurableJob-owned pool |
| Yjs maintenance scheduling and retry | Core | Core worker + Core-owned outbox hint |
| Yjs maintenance operation | YjsWorker | Core↔YjsWorker Kafka |
| RoutineTask lifecycle hint | DurableJob | DurableJob→RealtimeGateway Kafka |

RoutineTask claim, assignment, execution, business mutation, and lifecycle
finalization all happen in DurableJob through its own PostgreSQL connection.
Completion also writes the shared outbox row in the same database flow, so Core
does not consume a DurableJob result event.

Kafka and the Core outbox are at-least-once. The consumer platform cannot
atomically mark an arbitrary runtime side effect as complete, so it must not
pretend to own a global deduplication table. Each concrete handler owns its
idempotency boundary: it uses `eventId` (or the event's monotonic aggregate
version when a future contract defines one) in the same transaction or
idempotent store operation as its side effect. A redelivered event must then be
a no-op and still return success. RealtimeGateway's Redis relay may redeliver,
but local detach closes an existing socket or removes an existing channel; a
repeated revoke finds neither and is therefore a no-op. NOT-35 applies this rule to RealtimeGateway;
future Core, DurableJob, and Email consumers apply it in their own data
boundaries.

Handlers return `*kafka.ConsumerError` to classify a failure:

| Classification | Handling |
| --- | --- |
| `Transient` (the default for ordinary errors) | Retry locally with bounded exponential backoff. After `KAFKA_CONSUMER_MAXIMUM_ATTEMPTS`, write to DLQ. |
| `PoisonMessage` | Write to DLQ immediately; retrying cannot make the handler accept it. |
| `SchemaIncompatible` | Write to DLQ immediately. Envelope decode/contract validation produces this classification automatically. |

The retry limit is intentionally per active delivery lease. A crash or
rebalance before a terminal commit restarts delivery, preserving data rather
than silently discarding it. DLQ delivery contains the original topic,
partition, offset, key, unmodified source bytes, event ID when decodable,
classification, attempt count, and failure timestamp.

The initial dead-letter topic is `notegic.core.lifecycle.v1.dlq`, configured
with 30-day delete retention. After the cause has been corrected, operators
inspect the DLQ record, verify its original envelope is now compatible, and
re-publish its preserved `key` and `value` to `sourceTopic`. This re-drive goes
through the normal consumer/idempotency path; no consumer offsets are edited
manually. Poison or incompatible records should be corrected at the producer
or contract layer before re-drive.

Consumer telemetry uses `kafka.consume.*`, `kafka.consumer.lag`,
`kafka.retry.count`, and `kafka.dlq.count`, alongside structured failure logs.
`traceParent`/`traceState` in the event envelope are extracted before the
handler runs, so handler spans continue the producer trace.

### Core/Email direction map

Core authentication workflows enqueue one of the operation-specific
`SendWelcomeEmailRequestDto`, `SendValidationEmailRequestDto`, or
`SendSecurityAlertEmailRequestDto` DTOs in Core's transactional outbox.
`OutboxRelay` publishes the envelope to
`notegic.core.email.request.v1`; Email consumes it with the
`notegic-email-core-v1` group and hands the operation to its local email worker
manager. Email owns SMTP, templates, retries, and delivery-side failures. The
Core process no longer calls Email's business HTTP endpoints; Email's HTTP
transport now exposes only `/startedz` and `/healthz`.
The event contains the selected operation and its operation-specific DTO, not
access tokens, cookies, or database records. Invalid events are sent to the
consumer's DLQ and transient worker failures use bounded consumer retries.
