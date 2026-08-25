# Core v1 contracts

Core owns the versioned contracts for the business capabilities it provides.
The folders are organized by transport concern and then by business domain:

- `api/` contains Core operation RequestDto/ResponseDto contracts. Each domain
  has one folder and operation-oriented files such as `get.go`, `create.go`,
  `update.go`, and `search.go`.
- `events/` contains Core lifecycle Kafka event schemas. Core writes these
  events through its transactional outbox; consumers such as RealtimeGateway
  depend on the schema but do not own it.
- `graphql/` contains the Core GraphQL schema, fragments, client documents,
  generated Go artifacts, generated models, and scalars. Gateway owns GraphQL
  execution, but Core owns the schema and business-facing query contract.
- `types/` contains Core-only reusable DTO shapes shared by more than one Core
  API operation.

Cross-runtime data vocabulary is outside this versioned service contract:

- `contracts/types/enums/` owns enum values. Core and DurableJob database
  wrappers add persistence behavior without redefining values.
+ `contracts/types/blocknote/editable_block.go` owns the single recursive
  `ArborizedEditableBlock` input shape and the single
  `RawFlattenedEditableBlock` projection/persistence shape.
- `contracts/types/*_routine_task_payload.go` owns payloads used by both Core
  and DurableJob.

Gateway owns the private HTTP `Request[D]` / `Response[D]` envelope in
`contracts/gateway/v1`. It carries version, operation, request metadata, and
the Core-owned DTO. Delegation credentials stay in the authorization header;
browser JWTs and contexts are never serialized into either contract.
