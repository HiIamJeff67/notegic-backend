# Architecture and Ownership

The target structure and staged ownership are defined in
[microservice-architecture.md](../codebase-design/microservice-architecture.md). Do
not create empty packages just to satisfy the target tree; move an owner only
when the corresponding migration issue owns its implementation.

## Public HTTP and internal transport

Runtime status transports expose two probe endpoints with distinct meanings:
`/startedz` maps to the application `IsHealthy()` startup flag and only confirms
that the runtime has started its HTTP server; `/healthz` maps to `IsReady()` and
confirms that the runtime can accept its normal external operations. A status
router must only inspect its own runtime and must not probe another runtime.
Docker Compose `service_healthy` checks `/healthz`; dependency ordering remains
the responsibility of Compose or the deployment orchestrator.

```text
public HTTP: route -> binder -> controller -> Gateway Core adapter -> Core Gateway endpoint -> Core service -> repository/scope
GraphQL: Gateway executor -> resolver/dataloader -> Gateway Core adapter -> Core Gateway endpoint -> Core service
```

| Layer | Responsibility | Must not do |
| --- | --- | --- |
| Gateway route | Register URL, middleware, trace/metric names, and route permission policy | Parse requests, execute business workflow, query SQL |
| Gateway binder | Bind URI/query/header/body and validate the public HTTP contract before invoking its controller | Decide ownership, open domain transactions, query repositories |
| Gateway controller | Call an adapter with a validated request DTO and render a client-safe response | Bind public HTTP input or query repositories directly |
| Gateway Core adapter | Map a versioned contract to an outbound Core request and map the result/error back | Implement Core business rules or access Core data |
| Core Gateway endpoint | Verify delegation credential, map contract request/response, call Core service, and render the internal HTTP response | Re-apply public route semantics or query GORM directly |
| Core service | Validate workflow request, coordinate transaction and application dependencies | Render Gin/HTTP response |
| Repository/scope | Assemble persistence, permission, preload, soft-delete, and locking query | Import transport request/response or return HTTP status |

Gateway binders own public transport parsing and validation. They live in
`runtimes/clientgateway/transports/api/binders/`, alongside controllers. Controllers
receive validated request DTOs, invoke a Gateway adapter, and turn the result
into the client response. They do not contain public HTTP parsing, domain rules,
or data-access code.

Gateway cryptographically parses browser credentials and creates a short-lived
delegation credential for an internal request. The credential carries the calling
component, optional authenticated public user subject, trace/request identity, and
route-declared allowed permissions. Core verifies it, validates the forwarded
browser credential, then resolves its own trusted identity before passing
permissions to service/repository options. Browser JWTs and `context.Context` are
not DTO fields or internal request types.

`AllowedPermissions` is optional delegation metadata, not a second authentication
requirement. Routes that only need an authenticated user (for example, `users/me`
or authenticated auth operations) use `CallSecurly` without mounting an allowed
permission middleware. Routes that operate on a protected resource must mount an
explicit `AllowedPermissions...` middleware before `CallSecurly`; Core services
that require resource authorization read the strict permission context and fail
closed when that route policy is missing.

The client-facing ClientGateway transport lives in
`runtimes/clientgateway/transports/api/`. It owns routes, binders, controllers, and
client-only middlewares/interceptors. Reusable Gin cookie handlers live in
`shared/cookies/`.
`controller_func.go` defines the shared controller function type; binder packages
must import it explicitly as `apitransport`.

Gateway-to-microservice internal HTTP is separate at
`runtimes/clientgateway/transports/core/`: its `adapters/` package is the
outbound client boundary. If a Core service concern later needs middleware or
interceptors, it belongs under this transport, never under `api/`. Core's
inbound Gateway transport is
organized by responsibility:

```text
runtimes/core/transports/
  middlewares/              # delegation, authentication, role, and plan checks
  gateway/
    endpoints/              # endpoint interfaces, handlers, and adjacent tests
    routers/                # HTTP route registration and adjacent router tests
```

APIGateway is a separate runtime in `runtimes/apigateway/`. It owns its
API-key entry point, Core adapter, Redis cache, rate limits, configuration,
transports, and tests; it does not import ClientGateway source.

An endpoint owns parsing the delegated envelope, invoking a Core service, and
serializing the internal response. A router only constructs groups and binds
HTTP method/path to endpoint methods. Core middleware performs delegation and
Core-owned authorization before an endpoint runs. Keep `endpoint.go` only for
endpoint-wide helpers; render a successful operation response in its endpoint
method.

Kafka is also a transport boundary. Runtime-specific Kafka consumers and
producers belong under the transport owned by the peer they communicate with;
they do not belong in a generic `workers/` package:

```text
runtimes/core/transports/
  durablejob/
    consumers/
    eventbuilders/
    producers/
  yjsworker/
    consumers/
    producers/
  outbox_relay.go

runtimes/durablejob/transports/
  realtimegateway/
    eventbuilders/
    producers/
  yjsworker/
```

DurableJob has no Core-facing request/response Kafka transport. Its lifecycle
notifications to RealtimeGateway belong under
`runtimes/durablejob/transports/realtimegateway/`. A file under `producers/` is
a broker producer only when it owns a `Produce()` method and receives the
platform Kafka producer. A file under `eventbuilders/` only builds a versioned
envelope through `Build()`; it never publishes to Kafka. Core-owned events
written inside a Core database transaction use the Core `OutboxRelay` as their
Kafka publisher; DurableJob-owned lifecycle notifications use the
DurableJob-to-RealtimeGateway producer.

The transport owns Kafka envelopes, producer/consumer setup, retry and offset
handling, and calls the local service or engine through constructor-injected
dependencies. Scheduling and execution policy remain in the owning runtime's
service or engine; those packages must not import their transport back.

Kafka topic creation policy follows the same ownership boundary. Define one
explicit `TopicSpec` factory per queue under `shared/platform/kafka/topics/` and
wire those factories through `All()`. `TopicSpec` has no implicit defaults:
partitions, replication, retention, cleanup policy, in-sync replicas, and
dead-letter retention must be specified by the topic owner. The platform
provisioner validates and applies the spec but never owns queue policy or
global topic variables.

## Runtime-owned workers

Long-lived background loops belong to the owning runtime's `workers/` package,
not to `services/` or a generic `workers/` package under `shared/platform`:

```text
runtimes/<runtime>/
  services/   # request-scoped business workflows
  workers/    # runtime-owned long-lived loops and reconciliation
```

A runtime-owned worker must:

- define an `XxxWorkerInterface` before its concrete worker when the composition
  root or tests need a replaceable lifecycle boundary;
- define an `XxxWorker` struct and `NewXxxWorker(...)` constructor for explicit
  dependency injection;
- expose `Start(context.Context) func()` and use the returned function for
  cancellation and graceful shutdown;
- accept context cancellation, avoid reading environment variables, and never
  register HTTP routes or render HTTP responses;
- own scheduling concerns such as tickers, bounded scans, retries, and
  reconciliation triggers while delegating business mutations to injected
  services/repositories;
- be constructed and started only by the runtime's application composition root.

Workers are runtime infrastructure with business-domain awareness. They may
coordinate the owning runtime's data and service dependencies, but they must not
import another runtime's source package or become a hidden replacement for a
service method. A service method must remain directly callable without starting
the worker.

Core services are grouped by business ownership so a future runtime split has a
stable package boundary:

```text
runtimes/core/services/
  auth/                               # auth and OAuth
  user/                               # user, account, info, settings, billing plans
  shelves/                            # root shelf, sub shelf, item
  blocks/                             # block pack, block, Yjs persistence
  material/
  routines/                           # station, routine, routine tag, RoutineTask
    parsers/                           # RoutineTask payload validation
  other/                              # badge and theme
  realtime/
```

Core's RoutineTask service owns task lifecycle APIs and validates the versioned
payload before persisting a task. Core's RoutineTaskDependency service owns
single-relation and batch CRUD for dependency edges, including same-Routine and
cycle validation. Dependency mutations are persisted incrementally; replacing
an entire graph is not a public write operation. It does not contain RoutineTask execution
handlers, mutation dispatch, pattern resolvers, or completion application logic.
DurableJob owns assignment claiming, payload interpolation, permission checks,
CRUD execution for the four supported objects, per-item execution results, task
finalization, and completion event publication. Any future Core/DurableJob
request must cross an explicit transport boundary under `transports/`; Core
services must not import DurableJob execution packages. Block remains a
projection read model and must not gain RoutineTask append/update/reset mutation
methods in Core.

## Dependency direction

- Each runtime's `commands/` package may import its owning `runtimes/<runtime>/` packages.
- Gateway client/API and Core-adapter transport code may import contracts, shared,
  and its own code; it must not query Core data or import repositories/
  GORM schemas. RealtimeGateway does not construct Core services, query Core
  data, or synchronously call Core after ticket issuance; it communicates with
  YjsWorker and receives Core lifecycle facts through Kafka.
- A runtime may import contracts, shared, and its own data. A runtime
  must not import another service source package.
- Shared PostgreSQL schemas, table names, scopes, repository
  implementations, and shared repository inputs are importable platform code.
  Runtime-owned repository inputs remain private to that runtime. Runtime
  packages may compose these implementations in thin local wrappers to add
  runtime-specific methods; wrappers embed the shared concrete repository when
  no signature adaptation is needed. Their location does not grant database access:
  migration ownership remains in the owning runtime's manifest, and future
  PostgreSQL roles/grants must separately define which runtime can read or
  write each object.
- Core, Gateway, DurableJob, Email, and RealtimeGateway are separate Go
  environments. Each runtime owns a `go.mod` and `go.sum` beside its
  `application.go`; `contracts` and `shared` own their own module metadata as
  well. The repository root intentionally has no `go.mod` or `go.sum`.
  `go.work` is the root workspace manifest that composes these modules for
  local development, but it is not an application module. The independent
  `test` module owns integration, architecture, and cross-runtime test
  dependencies and may replace the local runtime modules. Do not add a runtime
  dependency merely to reuse another runtime's source. YjsWorker keeps its
  independent Node/TypeScript package environment.
- `shared` is the root-level cross-runtime utility layer. It may depend on
  contracts and the minimum common application support it genuinely needs;
  portable `shared/lib` never imports a Notegic package.
- `shared/cookies`, `shared/tokens`, and `contracts/types/exceptions` are shared semantic
  boundaries that remain at the root of `shared`; reusable implementation
  utilities belong under `shared/util/` (`editableblock`, `exceptionwriter`,
  and `responsewriter`). `shared/util` may use application-support packages,
  while `shared/lib` remains the stricter dependency-free library layer.
- PostgreSQL repository exceptions belong under
  `shared/platform/postgres/repositories/exceptions/` and are split into one file per domain;
  they are persistence-layer exceptions, not runtime service exception packages.
- Runtime-owned exception builders belong under each runtime's
  `runtimes/<runtime>/exceptions/`; PostgreSQL repository exception builders
  belong under `shared/platform/postgres/repositories/exceptions/`.
- The generic Kafka envelope is maintained in `contracts/types/events/`.
  Runtime event domains remain under their owning `contracts/<runtime>/v1/events/`
  package; email request payloads therefore live in
  `contracts/email/v1/events/`.
- `shared/platform` owns infrastructure mechanics, not User/Shelf/Routine
  business rules.
- Cross-runtime calls use a versioned contract and adapter/client. Core adapters
  are outbound only; a Core inbound transport is already the inbound adapter.

## PostgreSQL schema importability and migration ownership

All PostgreSQL GORM models that represent shared physical tables live under
`shared/platform/postgres/schemas/`, including their table names and relations.
Migration DDL is owned by the runtime that owns the tables and lives under that
runtime's `data/postgres/constraints/`, `data/postgres/triggers/`, or
`data/postgres/views/` directory. Runtimes import shared models directly; they
do not keep duplicate local schema definitions. All repository implementations,
inputs, and repository exceptions live under
`shared/platform/postgres/repositories/`, organized by domain. Constructors
receive the owning runtime's `*gorm.DB` pool; the shared package must not keep a
default database or import a runtime. Runtime-specific business workflows stay
in services/workers instead of duplicating repository implementations. All
reusable scopes live under `shared/platform/postgres/scopes/`. Runtime-only raw
query helpers belong under `runtimes/<service>/data/postgres/sqls/`; they are not
shared platform assets.

Each database-owning runtime keeps one migration manifest under its own
`data/postgres/` package. The manifest declares its owner and the tables, enums,
views, triggers, and constraints that it is responsible for migrating. A shared
`MigrationManifest` is only an ownership declaration and migration input; it
does not grant database access.

Database permissions use the shared `postgres.PermissionManifest` and the
`postgres.ApplyPermissions` reconciler. A deployment/admin command first uses
its admin connection to ensure the fixed role (`notegic_<runtime>`) exists,
pins the admin connection, and runs each migration phase in one transaction:
temporarily grant schema `CREATE`, migrate with `SET ROLE` as the owner, reset
the role, revoke `CREATE`, and commit. A failure or interruption rolls back the
whole phase, including the temporary grant. It then applies the permission
manifest and calls `postgres.VerifyPermissions` through the same admin
connection. Verification is read-only: it compares every declared non-owner
runtime grant, `PUBLIC` ACL, and owner/schema default ACL against the manifest
and fails the deployment with the affected runtime, privilege, and object when
they differ. Role bootstrap enforces `NOINHERIT`, removes role
memberships in both directions, revokes database/schema `PUBLIC` privileges,
and then grants the runtime only explicit `CONNECT` and schema `USAGE`. Object
reconciliation revokes `PUBLIC` and every existing runtime role before granting
only the privileges declared by the current manifest, so removed grants cannot
survive a later deployment. Manifests also declare current-database `CONNECT`,
schema access, required enums/sequences/functions, and default privileges for
objects an owner runtime creates; the current schemas use UUID identifiers, so
they declare no generated sequence grants. `gen_random_uuid()` is a PostgreSQL
builtin, not a runtime-owned `public` function, so it is not a manifest object.
An object appears only in its owner
runtime's manifest, including grants for other runtimes; consumer manifests
must not repeat a shared object's grants.
Compose places admin credentials only in one-shot database bootstrap/migration
services; runtime application startup opens only its own runtime pool. Triggers
and constraints are not independent permission objects; their authority follows
the table that owns them. Non-owner runtimes must not add another runtime's
migration manifest to their startup path.

## Composition roots

Environment configuration follows the same ownership rule. A runtime composition
root loads each typed owner config once from its `configs/` package and injects it into dependencies;
infrastructure config is colocated with its component at
`shared/platform/<component>/config.go`. Do not read environment variables
from transports, services, workers, clients, or middleware, and do not recreate
`shared/platform/config/`.

Keep runtime configuration files split by concern instead of placing every
loader and policy in `configs/config.go`: for example, Kafka consumer settings,
rate-limit policy, SMTP settings, renderer settings, cache TTLs, and outbox
relay settings each have an appropriately named file under the owning runtime's
`configs/` package. Delete empty configuration files; a configuration package
should expose only settings that are actually consumed by that runtime. The
package exposes one public `LoadConfig()` entry point; all concern-specific
loaders called by it use private `load...Config()` names. Shared infrastructure
loaders, such as PostgreSQL environment parsing, belong under
`shared/platform/<component>/` and are called by the runtime's private loader.

Runtime-owned Redis client sets and TTLs follow the same boundary. Each runtime
composition root creates its immutable Redis client set and injects it into the
runtime-owned cache stores. Cache clients must not read Redis topology or
database ranges from a global registry, and cache-specific expiry values remain
in the owning runtime config. The resulting clients are reused for that
runtime's stores, services, and transports instead of constructing a new client
inside an operation.

Do not introduce an application `modules/` package merely to wrap service
construction. The owning composition root constructs its scope -> repository ->
service dependencies directly, then passes each concrete service to the router or
runtime that uses it. Core's `NewCoreTransportRouter` is the composition root for
its inbound endpoints; the WebSocket runtime constructs only its own Gateway,
Core client, Redis lease store, and YjsWorker manager.

Router construction may instantiate endpoint objects from the services it
receives, but it must not recreate the services or conceal their dependencies in a
module wrapper. Constructor parameters, struct dependency fields, and constructor
assignments use the same order.

Every runtime-owned component receives its operational dependencies through its
constructor. This includes repositories, scopes, clients, generators, hashers,
and the component's runtime/domain exception instance. Do not call
`New...Repository()`, `New...Exception()`, or another dependency constructor
inside a service or operation method. Construct the dependency once in the
composition root, pass it to `New...Service()`/`New...Component()`, and store it
on the concrete struct. A component may construct a value that it owns directly
(for example, an immutable parser configuration), but it must not instantiate a
shared repository or a dependency owned by another package at the point of use.

Exception factories are also dependencies: a service that returns a domain
exception keeps its injected domain exception on the struct and uses that
instance throughout its methods. Distinct domains used by one service are
injected as distinct dependencies. Constructors do not silently replace a nil
dependency with a newly constructed shared repository or exception; the
composition root is responsible for providing the complete graph. Optional
dependencies must have an explicit no-op or nil contract documented by the
component.

Do not create a module, interface, adapter, or helper for a single anticipated
future use. A concrete boundary with a real caller is enough.

## Public operation ordering

One public operation has one ordering across route registration, binder
interface/implementation, controller interface/implementation, Gateway adapter,
Core Gateway endpoint interface/implementation, router registration, and Core service interface/implementation:

1. read: get one, get many/search, aggregate;
2. create: one then many;
3. update: one then many;
4. restore: one then many;
5. soft delete: one then many;
6. hard delete: one then many;
7. permission/sub-resource: get, create, update/upsert, delete;
8. visualization/chart;
9. GraphQL, system-only, or background operation.

Visualization/chart operations form their own family. GraphQL, system-only, and
background operations form another family when present. Do not casually reorder a
legacy file that is outside the operation group being changed.

## File and helper layout

Each controller, service, repository, or adapter file uses this order:

```text
package / imports
interface
concrete struct
constructor
optional auxiliary helpers (always declared before service methods)
public methods in interface order
optional visualization/chart methods
optional GraphQL/system-only methods
```

- Keep one blank line between top-level methods.
- In `runtimes/core/services/`, helper functions and helper methods belong at the top of the file, immediately after the constructor and before the public service methods. Keep the primary service workflows below the helpers; do not place a private helper after the method that calls it.
- Extract a helper only when two or more methods reuse the same named concept, or
  the inline logic would hide the primary workflow. One-call parsing, mapping,
  validation, temporary type, and wrapper variable stay inline.
- Use `sep30` only for GraphQL, system-only, and visualization method families.
  The separator must appear in both the service interface and the matching
  implementation, using the same semantic group. Do not use it for ordinary
  service methods, main methods, CRUD methods, permission methods, or helpers.
- Local struct/type declarations require a concrete domain name and repeated use
  within a complex query/result mapping. Do not create `Data`, `Result`, or
  `Params` wrappers for one handoff.

## GraphQL and background runtimes

GraphQL uses Scheme A: Gateway owns the executor, resolvers, and dataloaders.
GraphQL source SDL/fragments/documents, generated Go code, scalars, and generated
models live in `contracts/core/v1/graphql`. Generated files are regenerated from source and
never edited directly. GraphQL business
RequestDto/ResponseDto live in the same
`contracts/core/v1/api/<route-domain>/search.go`
as their owning Core service. Core exposes each GraphQL operation from that
service's endpoint and router; never create a shared GraphQLEndpoint or a central
Core GraphQL router.

DurableJob, Email, and YjsWorker own their own runtime, transport, and service-local
data/types/config. A Node/TypeScript runtime follows the same boundary rule as a
Go runtime: `transports/` owns external protocols, `services/` owns application
logic, `types/` owns runtime data shapes, and `configs/` owns runtime policy.
Core and every other runtime may add a runtime-owned
`workers/` package for long-lived background coordination. They must support
`context.Context` cancellation and graceful shutdown. Kafka, outbox, and consumer
reliability are separate Phase 3 concerns; Yjs update, awareness, and presence
stay out of the outbox.

## Minimal pre-change check

1. Locate the Gateway route/controller/adapter or Core Gateway endpoint/Core service
   owning the operation.
2. Locate the service data repository/scope and closest existing test.
3. Confirm the target dependency direction before adding an import.
4. Add a new package or boundary only for a real independently owned lifecycle or
   external contract.
