# DurableJob event contracts v1

This package owns the Kafka protocols coordinated by DurableJob:

- `RoutineTaskRunning` lifecycle hints sent directly to RealtimeGateway. They are
  transient execution hints emitted immediately before a DurableJob handler
  begins; a Kafka delivery failure does not cancel the claimed task.
Core owns Yjs maintenance scheduling and publishes its own hints. YjsWorker
owns the maintenance operation, command, and worker-result contracts in
`contracts/yjs-worker/v1/events`.

The generic envelope is imported from `contracts/types/events/`; this package owns the
topics, event types, and payloads. Consumer groups remain runtime deployment
configuration and are defined by the runtime transport composition roots. This
package does not import Core repositories, database schemas, or
Kafka clients.
