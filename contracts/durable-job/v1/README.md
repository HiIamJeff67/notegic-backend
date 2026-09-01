# DurableJob v1 contracts

This directory is the versioned boundary owned by the DurableJob service. A
caller uses these contracts when it invokes DurableJob; the path does not encode
the caller or transport direction.

Routine task execution uses the shared PostgreSQL instance directly:

```text
DurableJob -> PostgreSQL: claim, quota, execute, finalize

DurableJob claims tasks through its own PostgreSQL connection, executes the
runtime-owned handlers, and writes business mutations plus RoutineRecord and
RoutineTaskRecord state in the same database flow. RoutineTask remains the
immutable task definition during execution; Core is not involved in
routine-task execution.
```

`ClaimRoutinesRequestDto` is a capacity request, not a request for one
specific routine. DurableJob owns routine claiming, quota consumption, routine
records, and scheduling state. It returns one `RoutineAssignment` per claimed
Routine, with that Routine's claimed `RoutineTaskAssignment` values nested under
`routineTasks`.

RoutineTask dependencies must form an acyclic graph. Core rejects invalid
dependency changes at the API persistence boundary, and DurableJob validates
the claimed Routine snapshot again during Preparation. A failed validation
terminally blocks the current RoutineRecord and all unfinished
RoutineTaskRecords; it is eligible again only after the Routine definition is
changed.

The service owns its runtime, handlers, validation, and execution state. Its
claim persistence models are runtime-owned and are accessed through the shared
PostgreSQL instance. Realtime lifecycle notifications remain a separate
DurableJob-to-RealtimeGateway Kafka concern.
