# Microservice Architecture Migration

## Purpose

Notegic adopts a staged migration rather than a one-time domain/database split. `Core` remains the synchronous owner of the current user, shelf, block, routine, and station workflows until a later bounded-context decision proves a separate runtime and database are warranted.

## Target ownership

```text
runtimes/
 clientgateway/commands/        # ClientGateway command
 apigateway/commands/           # APIGateway command
 core/commands/
 realtimegateway/commands/
 durablejob/commands/
 email/commands/
contracts/
  client-gateway/v1/       # ClientGateway-owned public request/response envelope
  api-gateway/v1/          # APIGateway-owned public request/response contract
  core/
    v1/
      api/                 # Core-owned HTTP and GraphQL operation DTOs
      events/              # Core-owned Kafka lifecycle event contracts
      graphql/             # Core-owned GraphQL schema and generated artifacts
      types/               # Core-only reusable DTO shapes
  types/
    enums/                 # Canonical cross-runtime enum values
    editable_block.go      # Arborized input tree and raw flattened persistence shape
    *_routine_task_payload.go
  durablejob/v1/           # DurableJob-owned internal boundary
  email/v1/                # Email-owned internal boundary
  realtime-gateway/v1/     # RealtimeGateway-owned internal boundary
  yjs-worker/v1/           # YjsWorker-owned internal boundary
runtimes/
  clientgateway/                 # ClientGateway-owned edge implementation
    transports/
      api/
        binders/
        controllers/
        cookies/
        interceptors/
        middlewares/
        routes/
      core/
        adapters/
      realtimegateway/
        adapters/
  realtimegateway/           # standalone WebSocket edge runtime
    data/redis/              # Realtime leases and Realtime-specific rate limits
    transports/
      gateway/               # private API Gateway presence transport
      yjsworker/             # WebSocket client, connection/channel state, middleware
    workers/                 # YjsWorker connection manager
  core/
    data/
      postgres/             # Core composition wrappers, SQL, seeds, and migration manifest
      redis/                # Core-owned Redis caches and Lua libraries
      storage/              # Core-owned storage implementations
    services/
      routines/
        parsers/             # RoutineTask payload validation
        routine_task_dependency_service.go # RoutineTaskDependency lifecycle APIs
    workers/                # Core-owned long-lived reconciliation and background loops
  durablejob/               # independent runtime; direct shared PostgreSQL access
    data/postgres/           # DurableJob repository composition, and migration manifest
    services/routinetask/    # RoutineTask components
      preparation/           # Assignment payload preparation and interpolation
        preparers/            # Assignment preparation components
      plan/                  # DAG validation and deterministic object planning
        builders/             # Deterministic plan builders
      execution/             # CRUD execution components
      recovery/              # Optional stale execution lease recovery
        recoverers/           # Stale record recovery components
      analysis/              # Future post-execution analysis components
      *_service.go           # Phase orchestration services
    workers/routinetask/     # RoutineTask scheduling, execution, and result handoff loop
    transports/              # RealtimeGateway lifecycle and YjsWorker clients
  email/                    # independent runtime and SMTP sender
  yjsworker/                # standalone TypeScript runtime; no src/ layer
    configs/                # runtime tuning and environment-backed settings
    transports/             # Core HTTP/Kafka and Realtime WebSocket boundaries
      core/
        endpoints/
        routers/
        consumers/
        dispatchers/
      realtime/
      health/
    services/               # Yjs transformation and projection application logic
    types/                  # runtime data shapes; versioned contracts live in contracts/
shared/
  exceptions/                # base exception envelope and rendering helpers
  platform/                  # database, Redis, Kafka, and observability lifecycle
  cookies/                    # reusable Gin access/refresh cookie handlers
  tokens/                    # access, refresh, delegation, and realtime tokens
  util/
    editableblock/            # tree-to-row EditableBlock conversion
    exceptionwriter/          # public exception formatting
    responsewriter/           # shared HTTP response buffering
  constants/ types/ validations/
  lib/                       # portable libraries, including BlockNote schemas
docs/
infra/
tests/
```

Directories are created only when they receive an owned implementation. Runtime
composition roots and Cobra commands live under their owning runtime's `commands/`
directory (`runtimes/clientgateway/commands/`, `runtimes/apigateway/commands/`, `runtimes/core/commands/`, and so on). GraphQL
SDL, generated artifacts, and generator configuration live under
`contracts/core/v1/graphql/`. Later issues establish the remaining owners; no
empty directory tree is committed merely to mirror this diagram.

## Runtime module boundaries

Each Go runtime and each cross-runtime support layer owns an independent module:

```text
contracts/go.mod
shared/go.mod
runtimes/core/go.mod
runtimes/cli/go.mod
runtimes/clientgateway/go.mod
runtimes/apigateway/go.mod
runtimes/durablejob/go.mod
runtimes/email/go.mod
runtimes/realtimegateway/go.mod
test/go.mod
```

Each module has its own module metadata and can be tested or built from its
directory without inheriting another runtime's module metadata. `go.work` at
the repository root composes the modules for local development. The `test`
module is reserved for black-box integration and architecture tests and does
not import runtime production packages. Runtime E2E tests live under the
owning runtime's `test/` directory. A root-level `go test ./...` is therefore
not a supported command; use `make test-all` or `runtimes/cli`. DurableJob's
execution protocol is carried by versioned contracts; Core remains the sole
owner of database-backed task execution and state transitions.
Root integration tests under `test/integration/` use the committed
`infra/docker/docker-compose.integration.yaml` stack for PostgreSQL, Redis, and Kafka. Root
Makefile targets own the Compose lifecycle, while the test module only runs the
tests; this allows local developers to reuse an already-running stack. The
same Compose commands are used locally and in GitHub Actions, while
WebSocket connection load/soak and Kafka consumer-lag checks remain k6 scripts
under `test/load/`, invoked by the root Makefile and outside Go modules.
YjsWorker remains a separate Node/TypeScript environment. It follows the same
runtime layering as the Go services: `transports/` contains Core-facing HTTP and
Kafka boundaries plus the Realtime WebSocket boundary, `services/` contains Yjs
transformation logic, `types/` contains runtime shapes, and `configs/` contains
runtime policies. Protocol constants and their versioned TypeScript counterpart
are owned together by `contracts/yjs-worker/v1`; the planned Go-to-TypeScript
contract generator will keep the two definitions synchronized. Eviction and
batching values remain runtime policy and therefore stay in `configs/` rather
than the cross-runtime contract.

## Dependency direction

```text
runtime commands -> owning internal runtime
clientgateway -> contracts + shared + own edge implementation
apigateway -> contracts + shared + own edge implementation
clientgateway -X-> service data/repository
runtimes/<runtime>/* -> contracts + shared + own data
runtimes/<runtime>/* -X-> another runtime's source
shared/lib -X-> all Notegic packages
shared/platform -X-> domain business packages
```

`shared/lib` is portable: it may use the standard library and a necessary third-party library, but never imports a Notegic package. `shared/util` contains reusable application-facing utilities such as `editableblock`, `exceptionwriter`, and `responsewriter`; unlike `shared/lib`, these utilities may depend on shared application packages. Shared parsers return ordinary Go `error`; application boundaries map those errors to service exceptions.

## Shared contract types

`contracts/types/enums` is the canonical owner of cross-runtime enum values.
Core and DurableJob database enum wrappers import those values and add only the
PostgreSQL responsibilities (`Name`, `Scan`, `Value`, validation, and string
conversion); neither runtime redefines a value set.

`contracts/types/blocknote/editable_block.go` has exactly two EditableBlock shapes:
`ArborizedEditableBlock` is the recursive, validated transport tree, while
`RawFlattenedEditableBlock` is the persistence/projection row with parent and
sibling identifiers plus raw JSON props/content.
`shared/util/editableblock.FlattenEditableBlock` and
`shared/util/editableblock.FlattenEditableBlocks` convert the former to the latter. There are
no generic `ArborizedBlock`, `RawArborizedEditableBlock`, or duplicate
`FlattenedEditableBlock` types.

DurableJob-owned RoutineTask execution payloads live under
`contracts/durable-job/v1/types/routine-tasks/`. Core validates and persists the
protocol boundary while DurableJob handlers consume it without importing Core
source. Only genuinely portable structures, such as EditableBlock shapes,
remain under `contracts/types/`. BlockNote schemas remain portable in
`contracts/types/blocknote`.

Core-owned reusable DTO shapes under `contracts/core/v1/types/<domain>/` use
the package name `coretypes`. Each domain folder remains an independent Go
package; callers use explicit domain aliases only when multiple folders are
imported by one file.

## Public and internal transport

Client-facing HTTP uses `routes -> binders -> controllers -> Core service
adapter`. Binders own URI/query/header/body binding and transport validation;
controllers invoke an adapter and render the client-safe response. Client-only
cookies, middlewares, and interceptors remain under the same `api` transport.
Neither layer owns transactions, business workflows, or persistence queries.

Gateway internal-API adapters are outbound clients reusable by REST and
GraphQL. They live in
`runtimes/clientgateway/transports/core/adapters/`; client transport code must
not contain Core service client implementation. Core transports are
inbound adapters: they verify the internal delegation credential, map a
versioned request to a Core service call, and map the service result to a
versioned response. Core adapters exist only for Core outbound calls to another
runtime.

Core's inbound Gateway transport is organized as
`runtimes/core/transports/gateway/endpoints/` and
`runtimes/core/transports/gateway/routers/`. Endpoints own the
delegated request/service/response flow; routers only bind HTTP paths and
methods. Delegation and Core-owned authorization middleware belongs in
`runtimes/core/transports/gateway/middlewares/`. Tests live beside
their endpoint, router, or middleware target.

`runtimes/clientgateway/commands`, `runtimes/apigateway/commands`, `runtimes/core/commands`, and
`runtimes/realtimegateway/commands` are independent composition
roots. ClientGateway accepts browser HTTP and GraphQL traffic; APIGateway accepts
the allowlisted external API-key resource HTTP traffic; Core owns
PostgreSQL-backed operations and its private listener; RealtimeGateway is the
separate WebSocket edge runtime, directly addressed by the Nginx WebSocket
upstream. RealtimeGateway owns connection admission, ticket verification, Redis
leases, Realtime-specific rate limits, and long-lived Go-to-YjsWorker
connections. Its root `application.go` is the composition root: each `Start()`
creates a new WebSocket client with its own connection state; no
global singleton application instance is retained. It never constructs Core
repositories or services.

NOT-57 adds `runtimes/durablejob/commands` and `runtimes/email/commands` as independent
composition roots.
DurableJob is an independent process. It owns the RoutineTask claimer, its
PostgreSQL connection, quota consumption, task records, scheduling transitions,
local handlers, business mutations, and finalization. Its GORM models come from
`shared/platform/postgres/schemas`; it has no import path to Core's repositories,
scopes, or services. It publishes only lifecycle notifications to
RealtimeGateway through its own transport; routine-task persistence and
execution remain in DurableJob. Email owns its SMTP sender and queue and consumes
Core's versioned Kafka email request contract; its HTTP transport exposes only
started/health endpoints. Both
commands initialize observability and stop their workers/HTTP servers on
context cancellation or SIGTERM.

RoutineTask processing follows an explicit phase order:
`claimed -> preparation -> plan -> execution -> recovery (optional) -> analysis`. Each
phase owns its own validation and verification rules. A component belongs under
the earliest phase that executes it; pure phase-independent contracts and
dependency-graph validation may belong outside the phase directories when they
are intentionally shared by multiple phases. Phase orchestration services remain
directly under `services/routinetask/`.

Core owns RoutineTaskDependency relation CRUD and validates task membership,
duplicate edges, self-edges, cross-Routine edges, and cycles before persistence.
DurableJob consumes the persisted dependency relations while executing a
Routine; it does not expose an entire-graph replacement write path.

RealtimeGateway owns socket admission, tickets, leases, connection state, and
worker forwarding. Core owns authorization, durable Yjs state, and block
projection, but RealtimeGateway does not synchronously call Core for those
operations. YjsWorker sends its persistence and projection commands to Core
through Kafka; Core publishes lifecycle facts through its transactional outbox.

The Gateway extracts and cryptographically parses browser access/refresh tokens
without querying Core data, then applies route permission policy. It sends a
short-lived internal delegation credential containing the calling component as
`actor`, an authenticated user's public UUID as the optional `userSubject`,
request/trace identity, and allowed permissions. For secure calls it also
forwards browser cookies, User-Agent, and sanitized forwarding metadata through
the private client. Core validates both the delegation credential and the
forwarded browser token, then resolves its own internal user identity. Core
inbound routes use `DelegationMiddleware` for component-only operations and
`DelegationAuthenticatedMiddleware` when a user subject is required. Go
`context.Context` is never serialized as an internal transport contract.

The versioned Core `Response[D]` envelope carries only operation data. A
short-lived signed BlockPack channel ticket carries the room-admission and
document-quota policy snapshot required by the realtime runtimes: policy versions,
admission enforcement strategy, maximum subscribers, and maximum Block count.
RealtimeGateway verifies that ticket, applies atomic Redis lease admission, and
passes the attested Block quota to Yjs Worker. The worker acknowledges attach only
after the room is ready and enforces each complete update against a validation
Y.Doc before changing authoritative state.

## GraphQL

GraphQL follows Scheme A. The Gateway owns GraphQL execution, resolvers, and dataloaders. Resolvers call Core transports through Gateway adapters; they never access Core repositories, GORM schemas, or data packages directly.

GraphQL SDL, fragments, operation documents, generated Go artifacts, generated models, and scalar infrastructure belong in `contracts/core/v1/graphql/`. Generated files are never edited directly.

## Data cache ownership

Redis is divided into two explicit responsibilities: platform client lifecycle
and runtime-owned cache operations. A runtime owns an immutable Redis client set;
there is no cross-runtime cache registry or server-number lookup.

```mermaid
flowchart TB
  subgraph Platform["shared/platform/redis"]
    ClientSet["ClientSet<br/>immutable []*redis.Client"]
  end

  subgraph Core["Core runtime"]
    CoreStart["core.Start()"]
    UserRegister["userdata.Register(clientSet)"]
    UserStore["UserDataCacheStore<br/>runtime-owned"]
    UserClient["UserDataCacheClient"]
  end

  subgraph Gateway["Gateway runtime"]
    GatewayStart["gateway.Start()"]
    RateRegister["ratelimitrecord.Register(clientSet)"]
    RateStore["RateLimitRecordCacheStore<br/>runtime-owned"]
    RateClient["RateLimitRecordCacheClient"]
  end

  subgraph RealtimeGateway["RealtimeGateway runtime"]
    RealtimeGatewayStart["realtimegateway.Start()"]
    RealtimeRegisterGateway["realtimelease.Register(clientSet)"]
    LeaseStore["RealtimeLeaseCacheStore<br/>runtime-owned"]
    LeaseClient["RealtimeLeaseStore"]
  end

  CoreStart --> ClientSet --> UserRegister --> UserStore
  UserClient --> UserStore

  GatewayStart --> ClientSet --> RateRegister --> RateStore
  RateClient --> RateStore

  RealtimeGatewayStart --> ClientSet --> RealtimeRegisterGateway --> LeaseStore
  LeaseClient --> LeaseStore
```

### Platform lifecycle

`shared/platform/redis` owns no cache domain or business policy. Its `ClientSet`
creates and closes the Redis clients owned by one runtime. The set is immutable
after startup; cache clients choose a client by a private hash-based shard index
when a runtime has more than one client.

Each runtime-specific cache store is the initialization boundary for its own
Lua libraries and cache policy:

```go
store := userdata.Register(ctx, clientSet)
client := userdata.NewUserDataCacheClient(config, store)
```

Core, Gateway, and RealtimeGateway each create their own process-local
`ClientSet`. They may target the same infrastructure Redis during development,
but no runtime knows another runtime's Redis topology or server numbers.

### Runtime ownership

| Runtime | Registration | Redis ownership |
|---|---|---|
| Core | `userdata.Register(clientSet)` | User data and quota cache |
| Gateway API | `ratelimitrecord.Register(clientSet)` | API rate-limit records |
| RealtimeGateway | `realtimelease.Register(clientSet)` + `ratelimitrecord.Register(clientSet)` | Connection/channel leases, participant presence, Realtime rate-limit records, lifecycle Pub/Sub fan-out |
| DurableJob / Email | None currently | No registered Redis cache domain |

`UserDataCacheStore`, `RateLimitRecordCacheStore`, and
`RealtimeLeaseCacheStore` each own a runtime-injected `ClientSet` and their
cache library loading. Their matching `*CacheClient` types own key formatting,
shard selection, and cache operations; they never retrieve another runtime's
store from a global registry.

Core quota functions are one-function-per-file under
`runtimes/core/data/redis/userdata/libraries/`. The UserData store
embeds and joins them into one `user_quota_library`, then performs a single
`FUNCTION LOAD REPLACE` during `Initialize()`.

### Monthly RoutineTask execution quota

RoutineTask payload cost remains an immutable task attribute, calculated by the
database trigger. Monthly execution consumption is separate from that payload
accounting: `UserQuotaTable` stores the actor-owned consumed cost units and its
cycle timestamps. `QuotaCycleWorker` is a Core-owned daily reconciliation loop;
it initializes missing quota rows from the active billing cycle (or account
creation for a free user) and resets rows whose `next_reset_at` has elapsed.
DurableJob claim atomically consumes the task creator's accumulated cost units
against the creator's current plan limit in the same transaction that changes a
task to `Running`. Consequently, sharing a RoutineTask does not charge the user
who happens to execute or view it, and an unavailable/late worker cannot make a
task execute beyond the monthly quota.

`GetMyAccountResponseDto.routineTaskCostUnitCount` remains temporarily for
client contract compatibility, but now reports the actor's monthly consumed
execution cost from `UserQuotaTable`. Clients must not use it to reject
RoutineTask creation or payload edits; quota is enforced only when DurableJob claims
an execution.

### Registration and operation flow

```text
core.Start()
  -> clientSet := redis.NewClientSet(config)
  -> userdata.Register(ctx, clientSet)
    -> store.Initialize()
      -> FUNCTION LOAD REPLACE user_quota_library
  -> UserDataCacheClient(config, store)

Auth/User service
  -> UserDataCacheClient.Get/Set/Update(...)
    -> hash identifier -> private client-set shard
    -> UserDataCacheStore.ClientSet()
```

`runtimes/realtimegateway/data/redis/realtimelease/` owns the RealtimeGateway
runtime's user
connection and BlockPack subscriber lease lifecycle, active lease inspection,
participant presence, and presence PubSub fanout. A Core-issued BlockPack
channel ticket carries signed admission and document-quota policy snapshots;
RealtimeGateway uses the admission claims directly during atomic admission and
forwards the Block quota claim to Yjs Worker during attach.
Core authorizes channel admission when it issues a signed ticket, but it does
not participate in RealtimeGateway subscribe or write boundaries and does not
read RealtimeGateway Redis or the live subscriber count during ownership or
plan workflows. RealtimeGateway applies the ticket's verified channel policy;
its Lua scripts are private to that cache domain.

RealtimeGateway owns the lease cache and produces participant snapshots and
deltas. Its lifecycle Kafka consumer publishes received revocations to its local
Redis Pub/Sub channel so every RealtimeGateway instance detaches only its own
matching connections and releases their leases. Clients request the ephemeral
participant snapshot directly from RealtimeGateway; API Gateway does not proxy
the request and Core never reads RealtimeGateway-owned lease/cache state. Core's
lifecycle path is outbox → Kafka → RealtimeGateway.

RealtimeGateway Kafka consumers are transport components, grouped by the runtime
they receive events from: Core consumers live under
`runtimes/realtimegateway/transports/core/consumers/`, and Notification consumers
live under `runtimes/realtimegateway/transports/notification/consumers/`. The
long-lived WebSocket protocol implementation is shared by all client-facing
events under `runtimes/realtimegateway/transports/websocket/`; it is not owned
by the Yjs worker transport. Yjs worker coordination remains an injected
dependency of that WebSocket adapter, while notifications, lifecycle events,
presence, and future user-facing events use the same connector and outbound
queue.

`runtimes/core/data/storage/` owns Core's storage implementation;
Gateway does not access it directly.

## Lifecycle event contracts

The runtime-neutral envelope is defined in
[`contracts/types/events/`](../../contracts/types/events/). It contains no topic, consumer
group, or business payload, so `shared/platform/kafka` can consume any
runtime's event without importing that runtime's contracts. Runtime-owned event
families are kept at their boundaries:

- [`contracts/core/v1/events`](../../contracts/core/v1/events/) contains Core
  lifecycle facts and policy decisions consumed by RealtimeGateway.
- [`contracts/durable-job/v1/events`](../../contracts/durable-job/v1/events/)
  contains DurableJob-to-RealtimeGateway lifecycle hints. RoutineTask
  persistence and execution results are written through PostgreSQL, not Kafka.
- [`contracts/core/v1/events`](../../contracts/core/v1/events/)
  contains Core-owned lifecycle facts and Yjs maintenance hints.
- [`contracts/yjs-worker/v1/events`](../../contracts/yjs-worker/v1/events/)
  contains YjsWorker command/reply transport metadata and maintenance
  operations.

Core owns the facts and policy values in its lifecycle messages.
RealtimeGateway consumes them to execute detach or future admission behavior,
but never derives business rules. The event envelope is independent of the
Kafka client, outbox schema, and consumer implementation so NOT-34, NOT-35,
and NOT-33 can evolve those owners without changing the semantic boundary.

## Exceptions and lifecycle

`contracts/types/exceptions` contains only the base exception envelope and origin
classification. Core domain factories stay in
`runtimes/core/exceptions/`; Gateway-only failures use the base envelope
at their use site. They return the base envelope but are never imported across a
Gateway/service boundary. Gateway transport owns client-safe HTTP exception
rendering. Portable shared libraries and parsers never import or return an
application exception.

Kafka, transactional outbox, consumer reliability, and cross-Gateway event
fanout are Phase 3 concerns. Core writes a service-owned outbox record in the
same transaction as a lifecycle mutation; WebSocket consumes the resulting
event and only executes the already decided detach/policy action. Yjs updates,
awareness, presence, and Redis leases remain ephemeral transport traffic and do
not enter the outbox.

The platform Kafka client owns broker configuration, TLS/SASL setup, broker
health checks, and common telemetry names. Core and RealtimeGateway may run
in degraded mode when Kafka is unavailable, but their `/healthz` endpoints stay
unhealthy until the runtime can accept normal operations. Local provisioning and operational steps
are documented in [Kafka Local Development](../runbooks/kafka-local-development.md).

Core's PostgreSQL-to-Kafka handoff is a Core-owned transactional outbox. Domain
transactions persist `OutboxEventTable` rows before committing; a Core worker
claims rows with `FOR UPDATE SKIP LOCKED`, publishes the versioned event
envelope, and marks acknowledgements. This deliberately provides at-least-once
delivery, with consumer idempotency owned separately. See the
[transactional outbox design](../system-design/transactional-outbox.md).
