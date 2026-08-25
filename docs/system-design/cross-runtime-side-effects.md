# Cross-runtime side effects

This document records the final Phase 3 decision for side effects that cross the
Core, Email, DurableJob, RealtimeGateway, and future AI runtimes. It is the
implementation boundary for NOT-37; it does not introduce speculative runtime
code before the owning issue is ready.

## Decision matrix

| Side effect | Owner | Transport | Consistency boundary | Decision |
| --- | --- | --- | --- | --- |
| Welcome, validation, and security-alert email | Core → Email | Kafka event from Core's PostgreSQL transactional outbox | Core domain transaction commits the event; Email is at-least-once and idempotent | Adopt |
| Object deletion and storage garbage collection | Core → storage provider | Future storage contract and outbox event | The object provider is eventually consistent; cleanup must be retryable and idempotent | Defer to Phase 5 / NOT-45 |
| Block projection, embeddings, vector indexing, tool-calling, and orchestration | Core → future AI/indexing runtime | Versioned domain events, to be designed with the first AI consumer | BlockTable remains a rebuildable Core read model; Yjs remains the source of truth | Future follow-up |
| Yjs persistence and maintenance | Core / YjsWorker | Core↔YjsWorker Kafka command/result plus Core-owned hint outbox | Core owns the database transaction; Core worker owns scheduling and bounded retry | Adopt |
| Realtime room policy, presence, lease, and revocation | Core → RealtimeGateway | Existing Core outbox and Kafka fan-out | RealtimeGateway owns live connection state and its Redis cache | Already covered by NOT-35, NOT-60, and NOT-61 |
| Rate limits, cache invalidation, metrics, traces, and logs | Runtime-local | Runtime-local mechanisms | No business transaction or cross-runtime event is required | Excluded |

## Email boundary

Core is the initiating runtime and the source of truth for the business
transaction. An authentication workflow writes an Email dispatch event to the
Core outbox in the same database transaction as the user/account mutation. The
Core outbox relay publishes the versioned event to Kafka. Email consumes the
event, renders the appropriate template, and submits it to SMTP through its
existing local worker.

The Email contract owns its event payload and topic under
`contracts/email/v1/events`. Consumers must use the event ID for idempotency,
bounded retries, and a dead-letter path. The migration removes the synchronous
Core → Email HTTP dependency; it must be implemented as a focused follow-up so
that every authentication transaction can be audited for atomic outbox writes.

## Storage boundary

Storage deletion is intentionally not part of the current Phase 3 migration.
DigitalOcean Spaces, object lifecycle, and garbage-collection policy are
tracked by NOT-45 in Phase 5. Until that contract exists, Core keeps its current
storage semantics and no speculative storage event is added to the outbox.

## AI and indexing boundary

The persisted Yjs document and updates remain the source of truth. The
projected BlockTable read model remains valuable for API queries, structural
search, and future AI ingestion because it is queryable and rebuildable. A
future indexing runtime may consume versioned projection or block-change events
and write embeddings, vector indexes, graphs, or tool metadata without placing
binary Yjs snapshots on Kafka. The event schema and replay policy belong in a
separate AI/indexing issue.

## Exclusions and acceptance

- Yjs persistence/maintenance, realtime fan-out, and room-state behavior are
  not duplicated here; their owning issues already define those contracts.
- No deferred storage or AI work changes existing request paths.
- The Email migration boundary, owner, consistency model, idempotency key,
  retry/DLQ behavior, and data-sensitivity rules are explicitly documented.
- The remaining implementation work is limited to the dedicated Email Kafka
  migration issue; NOT-37 itself is complete as the architecture decision
  record.
