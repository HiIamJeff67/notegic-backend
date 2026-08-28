# YjsWorker event contracts v1

This package owns the YjsWorker command/reply transport metadata and topic
names. The command and reply payload schemas remain in the parent
`contracts/yjs-worker/v1` package. The generic envelope is imported from
`contracts/types/events/`; no runtime implementation or Kafka client is imported.

This package owns the Core-to-YjsWorker command/reply contracts. Yjs maintenance
hints are a Core-owned contract consumed by the Yjs worker; maintenance state is
persisted directly in PostgreSQL and is not returned over Kafka. DurableJob
requests are separate contracts owned by their respective runtime.

Consumer groups are runtime deployment configuration and are not part of this
contract package.
