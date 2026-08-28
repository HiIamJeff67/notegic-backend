# Kafka Local Development

## Scope

Kafka is the Phase 3 transport for versioned cross-runtime events. Local Compose
runs one KRaft broker and provisions the complete versioned topic catalog. This is a local
and integration environment only; production Kafka topology, credentials, and
delivery automation remain outside this runbook.

The contract and production topic strategy are defined in
[Kafka Event Contracts](../system-design/kafka-event-contracts.md). The local
broker deliberately uses three partitions and replication factor one; the
production strategy is twelve partitions, replication factor three, and
`min.insync.replicas=2`.

## Start and inspect

Start the broker and idempotent topic provisioner:

```bash
docker compose -f docker-compose.yaml -f docker-compose.init.yaml \
  up -d notegic-kafka notegic-kafka-init
docker compose -f docker-compose.yaml -f docker-compose.init.yaml \
  logs --follow notegic-kafka-init
```

The same provisioner can be run from the host with:

```bash
make kafka-topics
```

Inspect a provisioned topic:

```bash
docker compose exec notegic-kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --topic notegic.core.lifecycle.v1
```

Containers connect to `notegic-kafka:9092`. A process started on the host uses
`127.0.0.1:9094` by default. The catalog in
`shared/platform/kafka/topics` is the source of truth for Core, DurableJob,
Email, Notification, and YjsWorker topics. Each catalog entry uses delete
cleanup, seven-day retention, three local partitions, replication factor one,
and a paired 30-day dead-letter topic.
Automatic topic creation is disabled, so
an absent topic is always a provisioning failure.

## Runtime configuration

`shared/platform/kafka/config.go` reads the infrastructure connection values.
Each runtime keeps its own consumer policy in its local config package. Core,
DurableJob, and RealtimeGateway establish Kafka clients at startup when they
own a producer or consumer.

| Variable | Default | Purpose |
| --- | --- | --- |
| `KAFKA_BROKERS` | `127.0.0.1:9094` | Comma-separated broker addresses. |
| `KAFKA_DIAL_TIMEOUT` | `3s` | Broker connection timeout as a Go duration. |
| `KAFKA_TLS_ENABLED` | `false` | Enable TLS with system roots or the file values below. |
| `KAFKA_TLS_CA_FILE` | `/run/secrets/kafka/ca.pem` | PEM CA file path inside the runtime container. |
| `KAFKA_TLS_CERT_FILE`, `KAFKA_TLS_KEY_FILE` | empty | Optional mTLS pair; both are required together and are mounted from `/run/secrets/kafka/`. |
| `KAFKA_TLS_SERVER_NAME` | empty | Optional TLS certificate server name override. |
| `KAFKA_SASL_MECHANISM` | empty | `PLAIN`, `SCRAM-SHA-256`, or `SCRAM-SHA-512`. |
| `KAFKA_SASL_USERNAME`, `KAFKA_SASL_PASSWORD` | empty | Required whenever SASL is configured. |
| `KAFKA_CONSUMER_MAXIMUM_ATTEMPTS` | `3` | Per-delivery handler attempts before a transient error enters DLQ. |
| `KAFKA_CONSUMER_INITIAL_RETRY_BACKOFF` | `250ms` | First consumer retry delay as a Go duration. |
| `KAFKA_CONSUMER_MAXIMUM_RETRY_BACKOFF` | `5s` | Cap for consumer exponential retry backoff as a Go duration. |
| `KAFKA_CONSUMER_MAXIMUM_POLL_RECORDS` | `100` | Bound on records held while rebalance is blocked. |

TLS/SASL settings are never committed as Compose secrets. Compose mounts the
host-side `./secrets/kafka/` directory read-only at `/run/secrets/kafka/` for
Kafka client runtimes. Place `ca.pem` there when TLS is enabled, and use
`client.crt`/`client.key` when mutual TLS is required. Production deployments
should provision that directory through the host secret manager or decrypt the
SOPS-managed artifact before starting Compose.

## Health and degraded behavior

Core and RealtimeGateway start even if the initial Kafka ping fails. They log a
degraded-mode warning; their `/startedz` reports that the process has started,
while `/healthz` returns `503` until the runtime can accept normal operations.
The current startup status is evaluated once; restart the runtime after the
required broker connection becomes available so `/healthz` can report healthy.

Do not treat degraded health as permission to lose lifecycle events: Core
persists them in PostgreSQL outbox rows before committing the associated domain
mutation. The relay publishes those rows when Kafka recovers, and
RealtimeGateway consumes them before fan-out to its own Redis Pub/Sub channel.

## Telemetry names

`shared/platform/kafka` owns the common OpenTelemetry metric names. The
outbox relay and consumers must record these instead of inventing runtime-local
variants:

| Metric | Use |
| --- | --- |
| `kafka.broker.ping.duration`, `kafka.broker.ping.count` | Startup/health broker reachability. |
| `kafka.publish.duration`, `kafka.publish.count`, `kafka.publish.failure.count` | Producer and outbox relay results. |
| `kafka.consume.duration`, `kafka.consume.count` | Successful consumer handling. |
| `kafka.consumer.lag` | Per topic/group lag observation. |
| `kafka.retry.count`, `kafka.dlq.count` | NOT-33 retry and dead-letter behavior. |
