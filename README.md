<a><img src="global/images/logo/header-image.png" alt="Notegic" /></a>

# Notegic Backend

Notegic Backend is the server-side application for Notegic. It is a modular,
multi-runtime Go backend with a dedicated TypeScript Yjs worker. The repository
is currently maintained as a proprietary project; it is not the old open-source
starter architecture.

## Architecture

The system separates public traffic, core business operations, realtime
connections, and asynchronous workers into independently runnable runtimes:

| Runtime | Responsibility |
| --- | --- |
| `runtimes/clientgateway` | Client-facing HTTP/GraphQL entry point, client cookies, request safety, and internal Core adapter calls. |
| `runtimes/apigateway` | Independent API-key HTTP entry point for external integrations, with its own Core adapter, Redis cache, rate limits, configuration, transports, and tests. |
| `runtimes/core` | PostgreSQL-backed business operations, authorization, repositories, cache ownership, and transactional outbox publishing. |
| `runtimes/realtimegateway` | Realtime/WebSocket gateway, ticket verification, connection admission, realtime leases, presence, and YjsWorker connections. |
| `runtimes/durablejob` | Durable-job consumers and scheduling strategies. Core remains the owner of database-backed task state. |
| `runtimes/email` | Email delivery runtime and message templates. |
| `runtimes/yjsworker` | TypeScript Yjs document, projection, persistence, and compaction worker. |

The main request paths are:

```text
Client HTTP/GraphQL
    -> ClientGateway
    -> Core
    -> PostgreSQL / Redis

External API-key HTTP
    -> APIGateway
    -> Core
    -> PostgreSQL / Redis

Client WebSocket
    -> RealtimeGateway
    -> YjsWorker

Core
    -> transactional outbox
    -> Kafka
    -> DurableJob / Email / Realtime consumers
```

ClientGateway and APIGateway are separate HTTP processes, while
RealtimeGateway is the WebSocket edge. Realtime traffic does not pass through
either HTTP gateway. Core owns business authorization and durable state;
RealtimeGateway owns connection state and realtime Redis data.

## Repository layout

```text
contracts/                         Versioned cross-runtime contracts
  core/v1/                          Core API, events, and GraphQL contracts
  durablejob/v1/                    DurableJob contracts
  email/v1/                         Email contracts
  client-gateway/v1/                ClientGateway public request/response envelope
  api-gateway/v1/                   APIGateway public request/response contract
  realtime-gateway/v1/              RealtimeGateway contracts
  yjs-worker/v1/                    YjsWorker contracts
  types/                            Portable shared contract shapes

runtimes/
  cli/                              Shared Cobra command runner
  clientgateway/                    ClientGateway runtime
  apigateway/                       APIGateway runtime
  core/                             Core runtime and data ownership
  realtimegateway/                  Realtime/WebSocket gateway runtime
  durablejob/                       Durable-job runtime
  email/                            Email runtime
  yjsworker/                         TypeScript Yjs runtime

shared/                             Cross-runtime platform and utilities
test/                               Architecture, integration, load, and soak tests
infra/                              Nginx, staging, observability, and deployment files
docs/                               Architecture, conventions, contracts, and runbooks
```

Each Go runtime has its own `go.mod`, Dockerfile, application composition root,
and Cobra commands. ClientGateway and APIGateway also own separate edge
transports, Core adapters, Redis caches, rate limits, configurations, and
tests; runtime dependencies cross only through `contracts` and `shared`.

## Local development

### Prerequisites

- Go `1.26.x`
- Docker and Docker Compose
- Node.js `22.x` and pnpm `11.x` for `runtimes/yjsworker`
- SOPS and age for encrypted environment files
- A local decrypted `.env` containing the required database, Redis, Kafka, token,
  OAuth, SMTP, and observability settings

### Start the development stack

```sh
make compose-up
make -C runtimes/core migrate
make -C runtimes/core seed
```

`make compose-up` configures SOPS automatically. It decrypts the development
artifact `.env.enc` into a temporary file and passes that file to
Compose. Raw `docker compose up` does not invoke SOPS and only reads `.env`.
For encrypted local environments, configure `.sops.yaml` with the intended age
recipients and use:

```sh
make compose-up
```

See the [environment secrets runbook](docs/runbooks/environment-secrets.md) for
onboarding, CI/CD, and key rotation.

`make compose-up` runs Compose detached, waits for services to become running
or healthy, and returns a failure status when they cannot become ready. Use
`docker compose ps` to inspect the result and `docker compose logs -f <service>`
to follow one service.

For a production-like local stack:

```sh
COMPOSE_PROJECT_NAME=notegic-prod-local \
	COMPOSE_FILE=infra/docker/docker-compose.prod.yaml \
	COMPOSE_ENCRYPTED_ENV_FILE=secrets/envs/.env.production.enc \
	make compose-up
```

The production-like stack is a local pre-deployment check; it does not replace
staging or production validation against managed infrastructure, real secrets,
TLS termination, DNS, resource limits, or network policies. See the
[Docker local development runbook](docs/runbooks/docker-local-development.md)
for the complete workflow and cleanup commands.

### Tests and quality checks

The root Makefile is the single command surface used by local development,
GitHub Actions, and Jenkins:

```sh
make ci-format
make ci-vet
make ci-unit
make ci-race
make ci-generated
make ci-containers
```

Docker-backed tests are intentionally separate from ordinary pull-request
checks:

```sh
make test-integration
make test-integration-kafka
make test-load-websocket
make test-soak-websocket
make test-load-kafka-lag
```

The root Compose targets manage the repository's Compose test stack
(`infra/docker/docker-compose.integration.yaml`). The test targets themselves only execute
tests, so they can also be run against an already-running local stack:

```sh
make compose-integration-up
make test-integration
make test-integration-kafka
make compose-integration-down
```

For a complete local lifecycle, use `make test-integration-managed`; it starts
the stack, runs both integration suites, and removes the stack even when a test
fails. CI uses the explicit `compose-integration-up` and
`compose-integration-down` targets so setup and cleanup remain visible in the
workflow logs.

### GraphQL contracts

GraphQL schema and generated artifacts are Core-owned:

```sh
make -C contracts gql-generate
make -C contracts gql-regenerate
```

Generated files are checked in and CI verifies that regeneration leaves a clean
working tree.

## CI/CD and Jenkins

GitHub Actions is the primary repository automation:

- `ci.yml` runs format, vet, unit, race, generated-contract, and container gates
  for pull requests, protected branches, and version tags.
- `integration.yml` runs scheduled or manually triggered PostgreSQL/Redis/Kafka
  integration verification.
- `staging.yml` promotes an immutable GHCR image tag on an approved staging
  self-hosted runner and checks `/startedz` and `/healthz` for every runtime.

Jenkins is an optional self-hosted pipeline executor, not another Notegic
runtime. The root `Jenkinsfile` deliberately calls the same Makefile targets as
GitHub Actions. It is useful when an organization needs an on-premise runner,
private network access, a separate approval system, or an existing Jenkins
credential store. Its staging stage promotes an already-built immutable GHCR
tag; it does not rebuild source on the staging host. Jenkins credentials provide
GHCR access and deployment secrets, while the repository contains no secret
values.

Staging commands are shared by both systems:

```sh
IMAGE_REGISTRY=ghcr.io/ORG/REPO IMAGE_TAG=TAG \
COMPOSE_ENV_FILE=/etc/notegic/staging.env make staging-deploy
make staging-smoke
```

Formal production rollout, migration rollback, and disaster recovery are
separate operational work and are not implied by a successful staging run.

## Documentation

- [Microservice architecture](docs/codebase-design/microservice-architecture.md)
- [CI/CD pipeline](docs/codebase-design/ci-cd-pipeline.md)
- [Project conventions](docs/conventions/README.md)
- [Kafka local development](docs/runbooks/kafka-local-development.md)
- [Transactional outbox](docs/system-design/transactional-outbox.md)
- [Realtime protocol](docs/system-design/realtime-protocol.md)

## Licensing

Project code is distributed under the Notegic proprietary license:

- `LICENSE.md` — English
- `LICENSE(tw).md` — Traditional Chinese

Third-party notices and license texts are preserved under `LICENSES/`.

<!-- DEVLOG:START -->
## Development log

This section is automatically maintained from recent Git history. Detailed intent belongs in commit messages and design documents.

### Recent snapshots

- [2026-09/2026-09-01](docs/devlogs/2026-09/2026-09-01.md)
- [2026-08/2026-08-30](docs/devlogs/2026-08/2026-08-30.md)
- [2026-08/2026-08-28](docs/devlogs/2026-08/2026-08-28.md)
- [2026-08/2026-08-26](docs/devlogs/2026-08/2026-08-26.md)
- [2026-08/2026-08-25](docs/devlogs/2026-08/2026-08-25.md)
<!-- DEVLOG:END -->
