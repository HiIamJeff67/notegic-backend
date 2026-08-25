# DurableJob v1 contracts

This directory is the versioned boundary owned by the DurableJob service. A
caller uses these contracts when it invokes DurableJob; the path does not encode
the caller or transport direction.

Routine task execution uses the shared PostgreSQL instance directly:

```text
DurableJob -> PostgreSQL: claim, quota, execute, finalize

DurableJob claims tasks through its own PostgreSQL connection, executes the
runtime-owned handlers, and writes business mutations plus RoutineTask and
RoutineTaskRecord state in the same database flow. Core is not involved in
routine-task execution.
```

`ClaimRoutineTasksRequestDto` is a capacity request, not a request for one
specific task. DurableJob owns task claiming, quota consumption, task records,
and scheduling state. It returns one `RoutineTaskAssignment` per task claimed
within the requested batch size.

The service owns its runtime, handlers, validation, and execution state. Its
claim persistence models are runtime-owned and are accessed through the shared
PostgreSQL instance. Realtime lifecycle notifications remain a separate
DurableJob-to-RealtimeGateway Kafka concern.
