# Runtime Configuration

Configuration is owned by the runtime or infrastructure component that consumes
it. Environment variables are only read by a typed `LoadConfig` function; the
runtime composition root loads each configuration once and injects it into
clients, workers, transports, and services.

## Ownership

| Owner | Config file | Examples |
| --- | --- | --- |
| Core PostgreSQL connection | `internal/core/configs/postgres.go` | `CORE_DB_HOST`, `CORE_DB_USER=notegic_core`, `CORE_DB_PASSWORD`, `CORE_DB_NAME`, `CORE_DB_PORT` |
| DurableJob PostgreSQL connection | `internal/durablejob/configs/postgres.go` | `DURABLEJOB_DB_HOST`, `DURABLEJOB_DB_USER=notegic_durablejob`, `DURABLEJOB_DB_PASSWORD`, `DURABLEJOB_DB_NAME`, `DURABLEJOB_DB_PORT` |
| Notification PostgreSQL connection | `internal/notification/configs/postgres.go` | `NOTIFICATION_DB_*` |
| PostgreSQL deployment/admin connection | runtime `commands/database_command.go` | `DB_ADMIN_*` for the main database and `NOTIFICATION_DB_ADMIN_*` for Notification |
| Redis connection | `shared/platform/redis/config.go` | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_INIT_DB` |
| Kafka connection and TLS | `shared/platform/kafka/config.go` | `KAFKA_BROKERS`, `KAFKA_DIAL_TIMEOUT`, `KAFKA_TLS_*`, `KAFKA_SASL_*` |
| OpenTelemetry SDK | `shared/platform/observability/config.go` | `OTEL_SERVICE_*`, `OTEL_EXPORTER_OTLP_GRPC_ENDPOINT` |
| ClientGateway | `internal/clientgateway/configs/` | `CLIENT_GATEWAY_LISTEN_ADDRESS`, legacy `GATEWAY_LISTEN_ADDRESS`, `CORE_BASE_URL` |
| APIGateway | `internal/apigateway/configs/` | `API_GATEWAY_LISTEN_ADDRESS`, `CORE_BASE_URL` |
| Core | `internal/core/configs/` | `CORE_LISTEN_ADDRESS`, `OAUTH_GOOGLE_*`, `STORAGE_KEY_SALT`, `OUTBOX_RELAY_*`, user-data cache TTL, quota-cycle worker interval, Yjs document initialization endpoint/timeout |
| DurableJob | `internal/durablejob/configs/` | `DURABLEJOB_LISTEN_ADDRESS`, runtime Kafka and Yjs document initialization settings |
| Email | `internal/email/configs/` | `EMAIL_LISTEN_ADDRESS`, `SMTP_*`, `NOTEGIC_OFFICIAL_*`, `KAFKA_*` consumer settings |
| RealtimeGateway | `internal/realtimegateway/configs/` | `REALTIME_GATEWAY_LISTEN_ADDRESS`, `REALTIME_ENABLED`, `YJS_WORKER_URLS` |

`shared/platform/config/` must not be recreated. A platform component owns
only its infrastructure connection configuration; runtime policy remains with
the runtime that uses it.

Each runtime's `configs/postgres.go` explicitly reads the environment variables
for its own connection and passes the five values to
`shared/platform/postgres.LoadConfig(host, user, password, name, port)`. The
shared loader validates values only; it does not derive environment-variable
names from a prefix. This allows a runtime such as Notification to add another
independent PostgreSQL connection with its own complete host, user, password,
name, and port fields.

Each runtime connection uses a fixed role derived from
`contracts/types.Runtime.RoleName()`: `notegic_core`, `notegic_durablejob`, or
`notegic_notification`. Runtime startup opens only that runtime connection.
Role bootstrap, migration, permission reconciliation, and Core seed execution
run through each runtime's one-shot `*-database-init` Compose service before
the runtime begins serving traffic. Each service invokes the existing Cobra
commands in order (`bootstrapDB`, `migrateDB`, and, for Core, `seedDB`) rather
than adding a wrapper command. The admin connection is never part of the
application pool or a long-running runtime container. Compose supplies
`DB_ADMIN_*` or `NOTIFICATION_DB_ADMIN_*` only to the one-shot
`*-database-init` services; runtime services receive only their own runtime
connection values.

PostgreSQL database names and roles use lowercase `snake_case`, for example
`notegic_db` and `notegic_notification`. Docker Compose service names,
container names, and internal hostnames remain lowercase `kebab-case`, for
example `notegic-db` and `notegic-notification-db`; they are DNS names rather
than PostgreSQL identifiers.

Core owns the Yjs maintenance strategy policy. Its composition root loads these
values through `internal/core/configs.LoadConfig` and injects one immutable
`YjsMaintenanceStrategyConfig` into the Core worker:

```dotenv
CORE_YJS_MAINTENANCE_MAXIMUM_PENDING_HINTS=1000
CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_BATCH=32
CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_WORKERS=8
CORE_YJS_MAINTENANCE_MAXIMUM_REQUEST_ATTEMPTS=3
```

## Canonical duration names

Duration values use Go duration strings. Do not introduce numeric unit suffixes
such as `_SECONDS`, `_MILLISECONDS`, or `_HOURS`.

```dotenv
KAFKA_DIAL_TIMEOUT=3s
CORE_CLIENT_TIMEOUT=10s
CORE_USER_DATA_CACHE_EXPIRES_IN=1h
CORE_USER_DATA_CACHE_MAX_ROTATION_RETRIES=5
YJS_DOCUMENT_INITIALIZATION_WORKER_TIMEOUT=30s
KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF=250ms
KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF=5s
OUTBOX_RELAY_POLL_INTERVAL=1s
OUTBOX_RELAY_CLAIM_TIMEOUT=30s
OUTBOX_RELAY_INITIAL_BACKOFF=1s
OUTBOX_RELAY_MAXIMUM_BACKOFF=1m
OUTBOX_RELAY_RETENTION=168h
OUTBOX_RELAY_CLEANUP_INTERVAL=1h
CORE_QUOTA_CYCLE_WORKER_INTERVAL=24h
```

All credentials, salts, passwords, client secrets, and SASL credentials are
secrets. They are injected by local development tooling, Compose, or the
CI/CD credential store or deployment host secure storage and must never be
logged or committed.

## Environment secret storage

The project uses SOPS with age for encrypted environment files. Development
uses `.env` and `.env.enc`; production, test, and staging use
`secrets/envs/.env.<environment>` and `secrets/envs/.env.<environment>.enc`,
along with `.sops.yaml`, plaintext environment files, and age identities, are
deployment/local artifacts and are ignored by Git. Transfer encrypted files
only through an approved private channel; never commit them to GitHub. Each
developer, CI system, Jenkins agent, and production host owns a separate age
identity.

New members generate their own identity and send only the public recipient to
the maintainer. The maintainer uses `make env-updatekeys` to add or remove
recipients and `make env-rotate` after a removal or compromise. CI/CD decrypts
only at runtime using its credential-store identity; deployment scripts remove
temporary plaintext files after Compose exits. See
`docs/runbooks/environment-secrets.md` for the complete onboarding and
rotation workflow.

Redis topology is runtime-owned. Each runtime composition root creates an
immutable `shared/platform/redis.ClientSet` and injects it into its cache stores
and clients. Cache clients may select a private shard by hashing a key, but no
server number or cross-runtime Redis registry is part of their configuration.
Core currently reads `CORE_USER_DATA_CACHE_EXPIRES_IN` and
`CORE_USER_DATA_CACHE_MAX_ROTATION_RETRIES`; Gateway and
RealtimeGateway keep their rate-limit policies in
`internal/<runtime>/configs/rate_limit.go`.
